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
	"context"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	storagev1alpha1 "github.com/koor-tech/koor-operator/api/v1alpha1"
	hc "github.com/mittwald/go-helm-client"
	hcmock "github.com/mittwald/go-helm-client/mock"
)

var _ = Describe("KoorCluster controller", func() {
	const (
		KoorClusterNamePrefix = "test-koorcluster-"
		KoorClusterNamespace  = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	var (
		mockCtrl       *gomock.Controller
		mockHelmClient *hcmock.MockClient
		reconciler     *KoorClusterReconciler
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockHelmClient = hcmock.NewMockClient(mockCtrl)
		reconciler = &KoorClusterReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
	})

	Context("When creating a KoorCluster", func() {
		It("Should install the operator and the cluster helm charts", func() {
			gomock.InOrder(
				mockHelmClient.EXPECT().AddOrUpdateChartRepo(gomock.Any()).Return(nil).Times(1),
				mockHelmClient.EXPECT().UpdateChartRepos().Return(nil).Times(1),
				mockHelmClient.EXPECT().InstallOrUpgradeChart(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
					DoAndReturn(func(ctx interface{}, chartSpec *hc.ChartSpec, opts interface{}) (interface{}, error) {
						Expect(chartSpec.ReleaseName).To(Equal(KoorClusterNamespace))
						return nil, nil
					}),
				mockHelmClient.EXPECT().InstallOrUpgradeChart(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
					DoAndReturn(func(ctx interface{}, chartSpec *hc.ChartSpec, opts interface{}) (interface{}, error) {
						Expect(chartSpec.ReleaseName).To(Equal(KoorClusterNamespace + "-cluster"))
						return nil, nil
					}),
			)

			By("By creating a new KoorCluster")
			ctx := context.Background()
			koorCluster := &storagev1alpha1.KoorCluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "storage.koor.tech/v1alpha1",
					Kind:       "KoorCluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: KoorClusterNamePrefix,
					Namespace:    KoorClusterNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, koorCluster)).To(Succeed())
			Expect(reconciler.reconcileNormal(ctx, koorCluster, mockHelmClient)).To(Succeed())
		})
	})

	Context("When finalizing a KoorCluster", func() {
		It("Should uninstall the operator and the cluster helm charts", func() {
			gomock.InOrder(
				mockHelmClient.EXPECT().UninstallReleaseByName(KoorClusterNamespace+"-cluster").Return(nil).Times(1),
				mockHelmClient.EXPECT().UninstallReleaseByName(KoorClusterNamespace).Return(nil).Times(1),
			)

			By("By creating a KoorCluster with Finalizer")
			ctx := context.Background()
			koorCluster := &storagev1alpha1.KoorCluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "storage.koor.tech/v1alpha1",
					Kind:       "KoorCluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: KoorClusterNamePrefix,
					Namespace:    KoorClusterNamespace,
					Finalizers:   []string{storagev1alpha1.KoorClusterFinalizerName},
				},
			}
			Expect(k8sClient.Create(ctx, koorCluster)).To(Succeed())
			Expect(reconciler.handleFinalizer(ctx, koorCluster, mockHelmClient)).To(Succeed())
		})
	})
})
