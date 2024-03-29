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
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	hc "github.com/mittwald/go-helm-client"
	hcmock "github.com/mittwald/go-helm-client/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	storagev1alpha1 "github.com/koor-tech/koor-operator/api/v1alpha1"
	"github.com/koor-tech/koor-operator/mocks"
	"github.com/koor-tech/koor-operator/utils"
)

var _ = Describe("KoorCluster controller", func() {
	const (
		KoorClusterNamePrefix = "test-koorcluster-"
		KoorClusterNamespace  = "default"
		KsdReleaseName        = "ksd-test"
		KsdClusterReleaseName = "ksd-cluster-test"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250

		ksdCurrentVersion  = "v1.11.0"
		cephCurrentVersion = "v17.2.5"
		ksdLatestVersion   = "v1.11.1"
		cephLatestVersion  = "v17.2.6"
		kubeVersion        = "1.27.3"
		defaultSchedule    = "0 0 * * *"
		newSchedule        = "1 0 * * *"
	)

	var (
		mockCtrl          *gomock.Controller
		mockHelmClient    *hcmock.MockClient
		reconciler        *KoorClusterReconciler
		mockVS            *mocks.MockVersionService
		mockCronsRegistry *mocks.MockCronRegistry
	)

	rookRelease := &release.Release{
		Chart: &chart.Chart{
			Values: map[string]any{
				"image": map[string]any{
					"tag": ksdCurrentVersion,
				},
			},
		},
	}

	clusterRelease := &release.Release{
		Chart: &chart.Chart{
			Values: map[string]any{
				"cephClusterSpec": map[string]any{
					"cephVersion": map[string]any{
						"image": "quai.io/ceph/ceph:" + cephCurrentVersion,
					},
				},
			},
		},
	}

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockHelmClient = hcmock.NewMockClient(mockCtrl)
		mockVS = mocks.NewMockVersionService(mockCtrl)
		mockCronsRegistry = mocks.NewMockCronRegistry(mockCtrl)
		reconciler = &KoorClusterReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
			crons:  mockCronsRegistry,
			vs:     mockVS,
		}
	})

	Context("When creating a KoorCluster", func() {
		It("Should update status and install the operator and the cluster helm charts", func() {
			gomock.InOrder(
				mockHelmClient.EXPECT().AddOrUpdateChartRepo(gomock.Any()).Return(nil),
				mockHelmClient.EXPECT().UpdateChartRepos().Return(nil),
				mockHelmClient.EXPECT().InstallOrUpgradeChart(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, chartSpec *hc.ChartSpec, opts interface{}) (interface{}, error) {
						Expect(chartSpec.ReleaseName).To(Equal(KsdReleaseName))
						return rookRelease, nil
					}),
				mockHelmClient.EXPECT().InstallOrUpgradeChart(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, chartSpec *hc.ChartSpec, opts interface{}) (interface{}, error) {
						Expect(chartSpec.ReleaseName).To(Equal(KsdClusterReleaseName))
						return clusterRelease, nil
					}),
			)

			internalFunc := func() {
				panic("This should not be called!")
			}
			kcname := KoorClusterNamePrefix + "create"
			jobName := fmt.Sprintf("notification/%s/%s", KoorClusterNamespace, kcname)

			gomock.InOrder(
				mockCronsRegistry.EXPECT().Get(jobName).Return("", false),
				mockCronsRegistry.EXPECT().Get(jobName).Return(defaultSchedule, true),
				mockCronsRegistry.EXPECT().Get(jobName).Return(defaultSchedule, true),
			)

			mockCronsRegistry.EXPECT().Add(jobName, defaultSchedule, gomock.Any()).
				DoAndReturn(func(_ string, _ string, cmd func()) error {
					internalFunc = cmd
					return nil
				})

			mockVS.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&storagev1alpha1.DetailedProductVersions{
					Ceph: &storagev1alpha1.DetailedVersion{
						Version: cephLatestVersion,
					},
					Ksd: &storagev1alpha1.DetailedVersion{
						Version: ksdLatestVersion,
					},
				}, nil)

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
						NodeInfo: core.NodeSystemInfo{
							KubeletVersion: kubeVersion,
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
			koorCluster := &storagev1alpha1.KoorCluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "storage.koor.tech/v1alpha1",
					Kind:       "KoorCluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      kcname,
					Namespace: KoorClusterNamespace,
				},
				Spec: storagev1alpha1.KoorClusterSpec{
					KsdReleaseName:        KsdReleaseName,
					KsdClusterReleaseName: KsdClusterReleaseName,
				},
			}
			Expect(k8sClient.Create(ctx, koorCluster)).To(Succeed())
			Expect(reconciler.reconcileNormal(ctx, koorCluster, mockHelmClient)).To(Succeed())

			By("Checking status after create")
			key := types.NamespacedName{Name: kcname, Namespace: KoorClusterNamespace}
			createdKoorCluster := &storagev1alpha1.KoorCluster{}

			Eventually(func() error {
				return k8sClient.Get(ctx, key, createdKoorCluster)
			}).Should(Succeed())

			Expect(createdKoorCluster.Status.TotalResources.Nodes.Equal(resource.MustParse("3"))).To(BeTrue())
			Expect(createdKoorCluster.Status.TotalResources.Cpu.Equal(resource.MustParse("12"))).To(BeTrue())
			Expect(createdKoorCluster.Status.TotalResources.Memory.Equal(resource.MustParse("60G"))).To(BeTrue())
			Expect(createdKoorCluster.Status.TotalResources.Storage.Equal(resource.MustParse("600G"))).To(BeTrue())
			Expect(createdKoorCluster.Status.MeetsMinimumResources).To(BeFalse())
			Expect(createdKoorCluster.Status.CurrentVersions.Kube).To(Equal(kubeVersion))
			Expect(createdKoorCluster.Status.CurrentVersions.KoorOperator).To(Equal(utils.OperatorVersion))
			Expect(createdKoorCluster.Status.CurrentVersions.Ksd).To(Equal(ksdCurrentVersion))
			Expect(createdKoorCluster.Status.CurrentVersions.Ceph).To(Equal(cephCurrentVersion))

			By("Checking status after running internal function")
			internalFunc()
			Eventually(func() error {
				return k8sClient.Get(ctx, key, createdKoorCluster)
			}).Should(Succeed())
			Expect(createdKoorCluster.Status.LatestVersions.Ksd.Version).To(Equal(ksdLatestVersion))
			Expect(createdKoorCluster.Status.LatestVersions.Ceph.Version).To(Equal(cephLatestVersion))

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
				mockHelmClient.EXPECT().AddOrUpdateChartRepo(gomock.Any()).Return(nil),
				mockHelmClient.EXPECT().UpdateChartRepos().Return(nil),
				mockHelmClient.EXPECT().InstallOrUpgradeChart(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, chartSpec *hc.ChartSpec, opts interface{}) (interface{}, error) {
						Expect(chartSpec.ReleaseName).To(Equal(KsdReleaseName))
						return rookRelease, nil
					}),
				mockHelmClient.EXPECT().InstallOrUpgradeChart(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, chartSpec *hc.ChartSpec, opts interface{}) (interface{}, error) {
						Expect(chartSpec.ReleaseName).To(Equal(KsdClusterReleaseName))
						return clusterRelease, nil
					}),
			)

			By("Checking status after adding nodes")
			Expect(reconciler.reconcileNormal(ctx, createdKoorCluster, mockHelmClient)).To(Succeed())
			afterNodeKoorCluster := &storagev1alpha1.KoorCluster{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, afterNodeKoorCluster)
				if err != nil {
					return false
				}
				return afterNodeKoorCluster.Status.TotalResources.Nodes.Equal(resource.MustParse("4"))
			}, "5s").Should(BeTrue())
			Expect(afterNodeKoorCluster.Status.TotalResources.Cpu.Equal(resource.MustParse("20"))).To(BeTrue())
			Expect(afterNodeKoorCluster.Status.TotalResources.Memory.Equal(resource.MustParse("100G"))).To(BeTrue())
			Expect(afterNodeKoorCluster.Status.TotalResources.Storage.Equal(resource.MustParse("1000G"))).To(BeTrue())
			Expect(afterNodeKoorCluster.Status.MeetsMinimumResources).To(BeTrue())

			By("Updating the notification schedule")
			afterNodeKoorCluster.Spec.UpgradeOptions.Schedule = newSchedule
			Expect(k8sClient.Update(ctx, afterNodeKoorCluster)).To(Succeed())

			gomock.InOrder(
				mockHelmClient.EXPECT().AddOrUpdateChartRepo(gomock.Any()).Return(nil),
				mockHelmClient.EXPECT().UpdateChartRepos().Return(nil),
				mockHelmClient.EXPECT().InstallOrUpgradeChart(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, chartSpec *hc.ChartSpec, opts interface{}) (interface{}, error) {
						Expect(chartSpec.ReleaseName).To(Equal(KsdReleaseName))
						return rookRelease, nil
					}),
				mockHelmClient.EXPECT().InstallOrUpgradeChart(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, chartSpec *hc.ChartSpec, opts interface{}) (interface{}, error) {
						Expect(chartSpec.ReleaseName).To(Equal(KsdClusterReleaseName))
						return clusterRelease, nil
					}),
			)

			gomock.InOrder(
				mockCronsRegistry.EXPECT().Remove(jobName).Return(nil),
				mockCronsRegistry.EXPECT().Add(jobName, newSchedule, gomock.Any()).
					DoAndReturn(func(_ string, _ string, cmd func()) error {
						internalFunc = cmd
						return nil
					}),
			)

			mockVS.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed"))

			By("Checking status after updating notification schedule")
			Expect(reconciler.reconcileNormal(ctx, afterNodeKoorCluster, mockHelmClient)).To(Succeed())
			updatedKoorCluster := &storagev1alpha1.KoorCluster{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, updatedKoorCluster)
				if err != nil {
					return false
				}
				return updatedKoorCluster.Spec.UpgradeOptions.Schedule == newSchedule
			}, "5s").Should(BeTrue())

			By("Checking status after running internal function")
			internalFunc()
			Eventually(func() error {
				return k8sClient.Get(ctx, key, updatedKoorCluster)
			}).Should(Succeed())
			Expect(updatedKoorCluster.Status.LatestVersions.Ksd.Version).To(Equal(ksdLatestVersion))
			Expect(updatedKoorCluster.Status.LatestVersions.Ceph.Version).To(Equal(cephLatestVersion))
		})
	})

	Context("When finalizing a KoorCluster", func() {
		It("Should uninstall the operator and the cluster helm charts", func() {
			gomock.InOrder(
				mockHelmClient.EXPECT().UninstallReleaseByName(KsdClusterReleaseName).Return(nil),
				mockHelmClient.EXPECT().UninstallReleaseByName(KsdReleaseName).Return(nil),
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
				Spec: storagev1alpha1.KoorClusterSpec{
					KsdReleaseName:        KsdReleaseName,
					KsdClusterReleaseName: KsdClusterReleaseName,
				},
			}
			Expect(k8sClient.Create(ctx, koorCluster)).To(Succeed())
			Expect(reconciler.handleFinalizer(ctx, koorCluster, mockHelmClient)).To(Succeed())
		})
	})
})
