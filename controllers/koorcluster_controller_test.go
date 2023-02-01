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

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/golang/mock/gomock"
	hc "github.com/mittwald/go-helm-client"
	hcmock "github.com/mittwald/go-helm-client/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storagev1alpha1 "github.com/koor-tech/koor-operator/api/v1alpha1"
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
		It("Should update status and install the operator and the cluster helm charts", func() {
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

			ctx := context.Background()

			By("Creating Nodes on the cluster")
			nodePrefix := "node-"
			nodes := []*core.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: nodePrefix,
					},
					Status: core.NodeStatus{
						Capacity: core.ResourceList{
							core.ResourceCPU:              resource.MustParse("4"),
							core.ResourceMemory:           resource.MustParse("10G"),
							core.ResourceEphemeralStorage: resource.MustParse("100G"),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: nodePrefix,
					},
					Status: core.NodeStatus{
						Capacity: core.ResourceList{
							core.ResourceCPU:              resource.MustParse("4"),
							core.ResourceMemory:           resource.MustParse("20G"),
							core.ResourceEphemeralStorage: resource.MustParse("200G"),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: nodePrefix,
					},
					Status: core.NodeStatus{
						Capacity: core.ResourceList{
							core.ResourceCPU:              resource.MustParse("4"),
							core.ResourceMemory:           resource.MustParse("30G"),
							core.ResourceEphemeralStorage: resource.MustParse("300G"),
						},
					},
				},
			}

			for _, node := range nodes {
				Expect(k8sClient.Create(ctx, node)).To(Succeed())
			}

			By("By creating a new KoorCluster")
			name := KoorClusterNamePrefix + "create"
			koorCluster := &storagev1alpha1.KoorCluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "storage.koor.tech/v1alpha1",
					Kind:       "KoorCluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: KoorClusterNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, koorCluster)).To(Succeed())
			Expect(reconciler.reconcileNormal(ctx, koorCluster, mockHelmClient)).To(Succeed())

			By("Checking status after create")
			key := types.NamespacedName{Name: name, Namespace: KoorClusterNamespace}
			createdKoorCluster := &storagev1alpha1.KoorCluster{}

			Eventually(func() error {
				return k8sClient.Get(ctx, key, createdKoorCluster)
			}).Should(Succeed())

			Expect(createdKoorCluster.Status.TotalResources.Nodes.Equal(resource.MustParse("3"))).To(BeTrue())
			Expect(createdKoorCluster.Status.TotalResources.Cpu.Equal(resource.MustParse("12"))).To(BeTrue())
			Expect(createdKoorCluster.Status.TotalResources.Memory.Equal(resource.MustParse("60G"))).To(BeTrue())
			Expect(createdKoorCluster.Status.TotalResources.Storage.Equal(resource.MustParse("600G"))).To(BeTrue())
			Expect(createdKoorCluster.Status.MeetsMinimumResources).To(BeFalse())

			By("Adding a new node")
			newNode := &core.Node{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: nodePrefix,
				},
				Status: core.NodeStatus{
					Capacity: core.ResourceList{
						core.ResourceCPU:              resource.MustParse("8"),
						core.ResourceMemory:           resource.MustParse("40G"),
						core.ResourceEphemeralStorage: resource.MustParse("400G"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, newNode)).To(Succeed())
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

			By("Checking status after update")
			Expect(reconciler.reconcileNormal(ctx, koorCluster, mockHelmClient)).To(Succeed())
			updatedKoorCluster := &storagev1alpha1.KoorCluster{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, updatedKoorCluster)
				if err != nil {
					return false
				}
				return updatedKoorCluster.Status.TotalResources.Nodes.Equal(resource.MustParse("4"))
			}, "5s").Should(BeTrue())
			Expect(updatedKoorCluster.Status.TotalResources.Cpu.Equal(resource.MustParse("20"))).To(BeTrue())
			Expect(updatedKoorCluster.Status.TotalResources.Memory.Equal(resource.MustParse("100G"))).To(BeTrue())
			Expect(updatedKoorCluster.Status.TotalResources.Storage.Equal(resource.MustParse("1000G"))).To(BeTrue())
			Expect(updatedKoorCluster.Status.MeetsMinimumResources).To(BeTrue())
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
