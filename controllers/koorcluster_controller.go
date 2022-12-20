/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"embed"
	"text/template"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Masterminds/sprig/v3"
	storagev1alpha1 "github.com/koor-tech/koor-operator/api/v1alpha1"
	hc "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
)

// KoorClusterReconciler reconciles a KoorCluster object
type KoorClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var TemplateFs embed.FS

//+kubebuilder:rbac:groups=storage.koor.tech,resources=koorclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.koor.tech,resources=koorclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.koor.tech,resources=koorclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KoorCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *KoorClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var koorCluster storagev1alpha1.KoorCluster
	if err := r.Get(ctx, req.NamespacedName, &koorCluster); err != nil {
		log.Error(err, "unable to fetch KoorCluster")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	helmClient, err := hc.New(&hc.Options{
		Namespace: koorCluster.Namespace,
		Debug:     true,
		Linting:   true,
	})

	if err != nil {
		log.Error(err, "Cannot create new helm client")
		return ctrl.Result{}, err
	}

	// Add koor-release repo
	// helm repo add koor-release https://charts.koor.tech/release
	chartRepo := repo.Entry{
		Name: "koor-release",
		URL:  "https://charts.koor.tech/release",
	}

	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		log.Error(err, "Cannot add koor-release repo")
		return ctrl.Result{}, err
	}

	if err := helmClient.UpdateChartRepos(); err != nil {
		log.Error(err, "Cannot update chart repos")
		return ctrl.Result{}, err
	}

	templates, err := template.New("").Funcs(sprig.TxtFuncMap()).ParseFS(TemplateFs, "utils/*")
	if err != nil {
		log.Error(err, "Cannot parse templates")
		return ctrl.Result{}, err
	}

	// Install rook operator
	// helm install --create-namespace --namespace koor-ceph koor-ceph koor-release/rook-ceph -f utils/operatorValues.yaml
	operatorBuffer := new(bytes.Buffer)
	err = templates.ExecuteTemplate(operatorBuffer, "operatorValues.yaml", koorCluster)
	if err != nil {
		log.Error(err, "Cannot execute operator template")
		return ctrl.Result{}, err
	}

	operatorChartSpec := hc.ChartSpec{
		ReleaseName:     koorCluster.Namespace,
		ChartName:       "koor-release/rook-ceph",
		Namespace:       koorCluster.Namespace,
		CreateNamespace: true,
		UpgradeCRDs:     true,
		ValuesYaml:      operatorBuffer.String(),
	}

	_, err = helmClient.InstallOrUpgradeChart(ctx, &operatorChartSpec, nil)
	if err != nil {
		log.Error(err, "Cannot install or upgrade operator chart")
		return ctrl.Result{}, err
	}

	// Install rook cluster
	// helm install --create-namespace --namespace koor-ceph koor-ceph-cluster \
	//     --set operatorNamespace=koor-ceph koor-release/rook-ceph-cluster -f values-override.yaml
	clusterBuffer := new(bytes.Buffer)
	err = templates.ExecuteTemplate(clusterBuffer, "clusterValues.yaml", koorCluster)
	if err != nil {
		log.Error(err, "Cannot execute cluster template")
		return ctrl.Result{}, err
	}

	clusterChartSpec := hc.ChartSpec{
		ReleaseName:     koorCluster.Namespace + "-cluster",
		ChartName:       "koor-release/rook-ceph-cluster",
		Namespace:       koorCluster.Namespace,
		CreateNamespace: true,
		UpgradeCRDs:     true,
		ValuesYaml:      clusterBuffer.String(),
	}

	_, err = helmClient.InstallOrUpgradeChart(ctx, &clusterChartSpec, nil)
	if err != nil {
		log.Error(err, "Cannot install or upgrade cluster chart")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KoorClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.KoorCluster{}).
		Complete(r)
}
