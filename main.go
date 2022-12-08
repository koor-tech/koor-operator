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

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/repo"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	storagev1alpha1 "github.com/koor-tech/koor-operator/api/v1alpha1"
	"github.com/koor-tech/koor-operator/controllers"
	hc "github.com/mittwald/go-helm-client"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

const (
	defaultNamespace   = "rook-ceph"
	operatorValuesFile = "utils/operatorValues.yaml"
	clusterValuesFile  = "utils/clusterValues.yaml"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	helmClient, err := hc.New(&hc.Options{
		Namespace: defaultNamespace,
		Debug:     true,
		Linting:   true,
	})

	if err != nil {
		log.Fatal(err)
	}

	// Add rook-release repo
	// helm repo add rook-release https://charts.rook.io/release
	chartRepo := repo.Entry{
		Name: "rook-release",
		URL:  "https://charts.rook.io/release",
	}

	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		log.Fatal(err)
	}

	if err := helmClient.UpdateChartRepos(); err != nil {
		log.Fatal(err)
	}

	// Install rook operator
	// helm install --create-namespace --namespace rook-ceph rook-ceph rook-release/rook-ceph -f utils/operatorValues.yaml
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operatorChartSpec := hc.ChartSpec{
		ReleaseName:     "rook-ceph",
		ChartName:       "rook-release/rook-ceph",
		Namespace:       defaultNamespace,
		CreateNamespace: true,
		UpgradeCRDs:     true,
		ValuesOptions: values.Options{
			ValueFiles: []string{operatorValuesFile},
		},
	}

	_, err = helmClient.InstallOrUpgradeChart(ctx, &operatorChartSpec, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Install rook cluster
	// helm install --create-namespace --namespace rook-ceph rook-ceph-cluster rook-release/rook-ceph-cluster -f utils/clusterValues.yaml
	clusterChartSpec := hc.ChartSpec{
		ReleaseName:     "rook-ceph-cluster",
		ChartName:       "rook-release/rook-ceph-cluster",
		Namespace:       defaultNamespace,
		CreateNamespace: true,
		UpgradeCRDs:     true,
		ValuesOptions: values.Options{
			ValueFiles:   []string{clusterValuesFile},
			Values:       []string{"operatorNamespace="+defaultNamespace},
		},
	}

	_, err = helmClient.InstallOrUpgradeChart(ctx, &clusterChartSpec, nil)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)

	//

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "14254cf5.koor.tech",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.KoorClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KoorCluster")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
