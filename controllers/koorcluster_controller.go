/*
Copyright 2023 Koor Technologies, Inc. All rights reserved.

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
	"fmt"
	"reflect"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/Masterminds/sprig/v3"
	"github.com/itchyny/gojq"
	hc "github.com/mittwald/go-helm-client"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"

	storagev1alpha1 "github.com/koor-tech/koor-operator/api/v1alpha1"
	"github.com/koor-tech/koor-operator/utils"
)

// KoorClusterReconciler reconciles a KoorCluster object
type KoorClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	crons  CronRegistry
}

func NewKoorClusterReconciler(mgr ctrl.Manager) *KoorClusterReconciler {
	return &KoorClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		crons:  NewCronRegistry(),
	}
}

//+kubebuilder:rbac:groups=storage.koor.tech,resources=koorclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.koor.tech,resources=koorclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.koor.tech,resources=koorclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=nodes/status,verbs=get
// Needed for helm to work in olm
//+kubebuilder:rbac:groups=*,resources=*,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *KoorClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	koorCluster := &storagev1alpha1.KoorCluster{}
	if err := r.Get(ctx, req.NamespacedName, koorCluster); err != nil {
		log.Error(err, "unable to fetch KoorCluster")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	restConfig, err := ctrl.GetConfig()
	if err != nil {
		log.Error(err, "Cannot get controller config")
		return ctrl.Result{}, err
	}

	helmClient, err := hc.NewClientFromRestConf(&hc.RestConfClientOptions{
		Options: &hc.Options{
			Namespace: koorCluster.Namespace,
			Debug:     true,
			Linting:   true,
		},
		RestConfig: restConfig,
	})

	if err != nil {
		log.Error(err, "Cannot create new helm client")
		return ctrl.Result{}, err
	}

	const finalizerName = storagev1alpha1.KoorClusterFinalizerName

	if koorCluster.IsBeingDeleted() {
		if err := r.handleFinalizer(ctx, koorCluster, helmClient); err != nil {
			log.Error(err, "Cannot handle finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(koorCluster, finalizerName) {
		controllerutil.AddFinalizer(koorCluster, finalizerName)
		if err := r.Update(ctx, koorCluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	err = r.reconcileNormal(ctx, koorCluster, helmClient)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *KoorClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.KoorCluster{}).
		Watches(
			&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(r.findKoorClusters),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(ce event.CreateEvent) bool {
					node, ok := ce.Object.(*corev1.Node)
					if !ok {
						return false
					}
					return len(node.Status.Capacity) != 0
				},
				UpdateFunc: func(ue event.UpdateEvent) bool {
					oldNode, ok := ue.ObjectOld.(*corev1.Node)
					if !ok {
						return false
					}
					newNode, ok := ue.ObjectNew.(*corev1.Node)
					if !ok {
						return false
					}
					return !reflect.DeepEqual(oldNode.Status.Capacity, newNode.Status.Capacity)
				},
				GenericFunc: func(ge event.GenericEvent) bool {
					return false
				},
			}),
		).
		Complete(r)
}

func (r *KoorClusterReconciler) findKoorClusters(_ client.Object) []reconcile.Request {
	koorClusterList := &storagev1alpha1.KoorClusterList{}
	if err := r.List(context.TODO(), koorClusterList); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(koorClusterList.Items))
	for i := range koorClusterList.Items {
		item := &koorClusterList.Items[i]
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *KoorClusterReconciler) reconcileNormal(ctx context.Context, koorCluster *storagev1alpha1.KoorCluster, helmClient hc.Client) error {
	log := log.FromContext(ctx)
	if err := r.reconcileResources(ctx, koorCluster); err != nil {
		return err
	}

	if err := r.reconcileHelm(ctx, koorCluster, helmClient); err != nil {
		return err
	}

	if err := r.reconcileNotification(ctx, koorCluster, &VersionServiceClient{}); err != nil {
		return err
	}

	if err := r.Status().Update(ctx, koorCluster); err != nil {
		log.Error(err, "Unable to update KoorCluster status")
		return err
	}

	return nil
}

func (r *KoorClusterReconciler) reconcileResources(ctx context.Context, koorCluster *storagev1alpha1.KoorCluster) error {
	log := log.FromContext(ctx)

	nodeList := &corev1.NodeList{}
	if err := r.List(ctx, nodeList); err != nil {
		log.Error(err, "unable to list Nodes")
		return err
	}

	resources := &koorCluster.Status.TotalResources
	resources.Nodes = resource.NewQuantity(int64(len(nodeList.Items)), resource.DecimalSI)
	resources.Storage = &resource.Quantity{}
	resources.Cpu = &resource.Quantity{}
	resources.Memory = &resource.Quantity{}

	// sum resources
	for idx := range nodeList.Items {
		capacity := &nodeList.Items[idx].Status.Capacity
		resources.Storage.Add(*capacity.StorageEphemeral())
		resources.Cpu.Add(*capacity.Cpu())
		resources.Memory.Add(*capacity.Memory())
	}
	koorCluster.Status.MeetsMinimumResources = resources.MeetsMinimum()
	if !koorCluster.Status.MeetsMinimumResources {
		log.Info("The cluster does not meet the minimum resource requirements")
		// TODO add event for minimum resources
	}

	return nil
}

func (r *KoorClusterReconciler) reconcileHelm(ctx context.Context, koorCluster *storagev1alpha1.KoorCluster, helmClient hc.Client) error {
	log := log.FromContext(ctx)
	// Add koor-release repo
	// helm repo add koor-release https://charts.koor.tech/release
	chartRepo := repo.Entry{
		Name: "koor-release",
		URL:  "https://charts.koor.tech/release",
	}

	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		log.Error(err, "Cannot add koor-release repo")
		return err
	}

	if err := helmClient.UpdateChartRepos(); err != nil {
		log.Error(err, "Cannot update chart repos")
		return err
	}

	templates, err := template.New("").Funcs(sprig.TxtFuncMap()).ParseFS(&utils.Templates, "*")
	if err != nil {
		log.Error(err, "Cannot parse templates")
		return err
	}

	// Install rook operator
	// helm install --create-namespace --namespace <namespace> <namespace>-rook-ceph koor-release/rook-ceph -f utils/operatorValues.yaml
	operatorBuffer := new(bytes.Buffer)
	err = templates.ExecuteTemplate(operatorBuffer, "operatorValues.yaml", koorCluster)
	if err != nil {
		log.Error(err, "Cannot execute operator template")
		return err
	}

	operatorChartSpec := hc.ChartSpec{
		ReleaseName:     koorCluster.Namespace + "-rook-ceph",
		ChartName:       "koor-release/rook-ceph",
		Namespace:       koorCluster.Namespace,
		CreateNamespace: true,
		UpgradeCRDs:     true,
		ValuesYaml:      operatorBuffer.String(),
	}

	operatorRelease, err := helmClient.InstallOrUpgradeChart(ctx, &operatorChartSpec, nil)
	if err != nil {
		log.Error(err, "Cannot install or upgrade operator chart")
		return err
	}

	rookVersion, err := getRookVersion(operatorRelease)
	if err != nil {
		log.Error(err, "Could not find rook version")
	} else {
		log.Info("Found rook version", "rookVersion", rookVersion)
		koorCluster.Status.CurrentVersions.Rook = rookVersion
	}

	// Install rook cluster
	// helm install --create-namespace --namespace <namespace> <namespace>-rook-ceph-cluster \
	//     --set operatorNamespace=<namespace> koor-release/rook-ceph-cluster -f utils/clusterValues.yaml
	clusterBuffer := new(bytes.Buffer)
	err = templates.ExecuteTemplate(clusterBuffer, "clusterValues.yaml", koorCluster)
	if err != nil {
		log.Error(err, "Cannot execute cluster template")
		return err
	}

	clusterChartSpec := hc.ChartSpec{
		ReleaseName:     koorCluster.Namespace + "-rook-ceph-cluster",
		ChartName:       "koor-release/rook-ceph-cluster",
		Namespace:       koorCluster.Namespace,
		CreateNamespace: true,
		UpgradeCRDs:     true,
		ValuesYaml:      clusterBuffer.String(),
	}

	clusterRelease, err := helmClient.InstallOrUpgradeChart(ctx, &clusterChartSpec, nil)
	if err != nil {
		log.Error(err, "Cannot install or upgrade cluster chart")
		return err
	}
	cephVersion, err := getCephVersion(clusterRelease)
	if err != nil {
		log.Error(err, "Could not find ceph version")
	} else {
		log.Info("Found ceph version", "cephVersion", cephVersion)
		koorCluster.Status.CurrentVersions.Ceph = cephVersion
	}

	return nil
}

func getRookVersion(rel *release.Release) (string, error) {
	result := ""
	query, err := gojq.Parse(".image.tag")
	if err != nil {
		// This should not happen
		return result, errors.Wrap(err, "Failed to compile rook query")
	}
	iter := query.Run(rel.Chart.Values)
	rookVersion, ok := iter.Next()
	if !ok {
		return result, fmt.Errorf("Could not find rook version via query")
	}

	result, ok = rookVersion.(string)
	if !ok {
		return result, fmt.Errorf("Found field is not a string")
	}
	return result, nil
}

func getCephVersion(rel *release.Release) (string, error) {
	result := ""
	query, err := gojq.Parse(".cephClusterSpec.cephVersion.image")
	if err != nil {
		// This should not happen
		return result, errors.Wrap(err, "Failed to compile ceph query")
	}

	iter := query.Run(rel.Chart.Values)
	cephImage, ok := iter.Next()
	if !ok {
		return result, fmt.Errorf("Could not find ceph image via query")
	}

	cephImageStr, ok := cephImage.(string)
	if !ok {
		return result, fmt.Errorf("Ceph image is not a string")
	}

	cephImageParts := strings.Split(cephImageStr, ":")
	if len(cephImageParts) != 2 {
		return result, fmt.Errorf("Ceph image is malformatted")
	}
	return cephImageParts[1], nil
}

func notificationJobName(koorCluster *storagev1alpha1.KoorCluster) string {
	jobName := "notification"
	nn := types.NamespacedName{
		Name:      koorCluster.Name,
		Namespace: koorCluster.Namespace,
	}
	return fmt.Sprintf("%s/%s", jobName, nn.String())
}

func (r *KoorClusterReconciler) reconcileNotification(ctx context.Context, koorCluster *storagev1alpha1.KoorCluster, vs VersionService) error {
	log := log.FromContext(ctx)
	jobName := notificationJobName(koorCluster)
	oldSchedule, ok := r.crons.Get(jobName)
	if !koorCluster.Spec.NotificationOptions.Enabled {
		if ok {
			// Notifications should be disabled
			r.crons.Remove(jobName)
		}
		return nil
	}

	newSchedule := koorCluster.Spec.NotificationOptions.Schedule
	if ok && newSchedule == oldSchedule {
		// Nothing changed
		return nil
	}

	if ok {
		log.Info("Schedule changed, remove old", "old", oldSchedule, "new", newSchedule)
		r.crons.Remove(jobName)
	}

	err := r.crons.Add(jobName, newSchedule, func() {
		currentKoorCluster := &storagev1alpha1.KoorCluster{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: koorCluster.Name, Namespace: koorCluster.Namespace}, currentKoorCluster)
		if k8serrors.IsNotFound(err) {
			log.Info("KoorCluster not found, deleting the job")
			r.crons.Remove(jobName)
			return
		}
		if err != nil {
			log.Error(err, "unable to fetch KoorCluster inside schedule")
			return
		}

		isUpdated := false
		latestCephVersion, err := vs.LatestCephVersion(currentKoorCluster.Spec.NotificationOptions.CephEndpoint)
		if err != nil {
			log.Error(err, "unable to find latest ceph version")
		} else {
			currentKoorCluster.Status.LatestVersions.Ceph = latestCephVersion
			isUpdated = true
		}

		latestRookVersion, err := vs.LatestRookVersion(currentKoorCluster.Spec.NotificationOptions.RookEndpoint)
		if err != nil {
			log.Error(err, "unable to find latest rook version")
		} else {
			currentKoorCluster.Status.LatestVersions.Rook = latestRookVersion
			isUpdated = true
		}

		// TODO add event if current version is not latest

		if isUpdated {
			if err := r.Status().Update(ctx, currentKoorCluster); err != nil {
				log.Error(err, "Unable to update KoorCluster status")
			}
		}
	})

	return err
}

// Handle finalizer and uninstall releases
func (r *KoorClusterReconciler) handleFinalizer(ctx context.Context, koorCluster *storagev1alpha1.KoorCluster, helmClient hc.Client) error {
	log := log.FromContext(ctx)
	if !controllerutil.ContainsFinalizer(koorCluster, storagev1alpha1.KoorClusterFinalizerName) {
		log.Info("No Finalizer")
		return nil
	}

	releaseName := koorCluster.Namespace + "-rook-ceph-cluster"
	if err := helmClient.UninstallReleaseByName(releaseName); err != nil {
		log.Error(err, "Failed to uninstall release", "releaseName", releaseName)
	}

	releaseName = koorCluster.Namespace + "-rook-ceph"
	if err := helmClient.UninstallReleaseByName(releaseName); err != nil {
		log.Error(err, "Failed to uninstall release", "releaseName", releaseName)
	}

	// remove our finalizer from the list and update it.
	controllerutil.RemoveFinalizer(koorCluster, storagev1alpha1.KoorClusterFinalizerName)
	return r.Update(ctx, koorCluster)
}
