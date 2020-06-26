/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */
package bookkeepercluster

import (
	"context"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bookkeeper cluster")
}

var _ = Describe("Bookkeeper Cluster Version Sync", func() {
	const (
		Name      = "example"
		Namespace = "default"
	)

	var (
		s = scheme.Scheme
		r *ReconcileBookkeeperCluster
	)

	var _ = Describe("Upgrade Test", func() {
		var (
			req reconcile.Request
			b   *v1alpha1.BookkeeperCluster
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			b = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			s.AddKnownTypes(v1alpha1.SchemeGroupVersion, b)
		})

		Context("Cluster condition prior to Upgrade", func() {
			var (
				client client.Client
				err    error
			)
			BeforeEach(func() {
				client = fake.NewFakeClient(b)
				r = &ReconcileBookkeeperCluster{client: client, scheme: s}
				_, err = r.Reconcile(req)
			})
			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("Initial status", func() {
				var (
					foundBookeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, err = r.Reconcile(req)
					foundBookeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookeeper)
				})

				It("should have current version set to spec version", func() {
					Ω(foundBookeeper.Status.CurrentVersion).Should(Equal(foundBookeeper.Spec.Version))
				})

				It("should set upgrade condition and status to be false", func() {
					_, upgradeCondition := foundBookeeper.Status.GetClusterCondition(v1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Status).Should(Equal(corev1.ConditionFalse))
				})
			})
			Context("syncClusterVersion when cluster in upgrading state", func() {
				var (
					err            error
					foundBookeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					foundBookeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookeeper)
					foundBookeeper.Status.SetUpgradingConditionTrue("UpgradeBookeeper", "0")
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.syncClusterVersion(foundBookeeper)
				})
				It("Error should be nil when the target version is Empty", func() {
					Ω(err).Should(BeNil())
				})
				It("Error should be nil when the target version is equal to current version", func() {
					foundBookeeper.Status.SetUpgradingConditionTrue("UpgradeBookeeper", "0")
					foundBookeeper.Status.TargetVersion = "0.6.1"
					foundBookeeper.Status.CurrentVersion = "0.6.1"
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.syncClusterVersion(foundBookeeper)
					Ω(err).Should(BeNil())
				})
				It("Error should be not nil when the target version is not equal to current version", func() {
					foundBookeeper.Status.SetUpgradingConditionTrue("UpgradeBookeeper", "0")
					foundBookeeper.Status.TargetVersion = "0.7.1"
					foundBookeeper.Status.CurrentVersion = "0.6.1"
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.syncClusterVersion(foundBookeeper)
					Ω(strings.ContainsAny(err.Error(), "failed to get statefulset ()")).Should(Equal(true))
				})
				It("Error should be nil when cluster is in rollbackfailedstate", func() {
					b.Status.SetErrorConditionTrue("RollbackFailed", " ")
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.syncClusterVersion(foundBookeeper)
					Ω(err).Should(BeNil())
				})
			})
			Context("syncClusterVersion when cluster in upgrade failed state", func() {
				var (
					err            error
					foundBookeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					foundBookeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookeeper)
					foundBookeeper.Status.SetErrorConditionTrue("UpgradeFailed", " ")
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.syncClusterVersion(foundBookeeper)
				})
				It("Error should be nil", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("syncClusterVersion when cluster in rollback failed state", func() {
				var (
					err            error
					foundBookeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					foundBookeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookeeper)
					foundBookeeper.Status.Init()
					foundBookeeper.Status.SetErrorConditionTrue("RollbackFailed", " ")
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.syncClusterVersion(foundBookeeper)
				})
				It("Error should be nil", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("syncClusterVersion when syncCompleted return true ", func() {
				var (
					err            error
					foundBookeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					foundBookeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookeeper)
					foundBookeeper.Status.TargetVersion = foundBookeeper.Spec.Version
					foundBookeeper.Status.CurrentVersion = "0.5.1"
					foundBookeeper.Status.Init()
					foundBookeeper.Status.SetUpgradingConditionTrue(" ", " ")
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.syncClusterVersion(foundBookeeper)
				})
				It("Error should be nil", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("rollbackClusterVersion", func() {
				var (
					err            error
					foundBookeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					foundBookeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookeeper)
					foundBookeeper.Status.UpdateProgress("UpgradeErrorReason", "")
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.rollbackClusterVersion(foundBookeeper, "0.6.1")
				})
				It("Error should not be nil", func() {
					Ω(strings.ContainsAny(err.Error(), "failed to get statefulset ()")).Should(Equal(true))
				})
			})
			Context("rollbackClusterVersion syncCompleted return true ", func() {
				var (
					err            error
					foundBookeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					foundBookeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookeeper)
					foundBookeeper.Status.TargetVersion = foundBookeeper.Spec.Version
					foundBookeeper.Status.Init()
					foundBookeeper.Status.UpdateProgress("UpgradeErrorReason", "")
					r.client.Update(context.TODO(), foundBookeeper)
					err = r.rollbackClusterVersion(foundBookeeper, "0.6.1")
				})
				It("Error should be nil", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("checkUpdatedPods", func() {
				var (
					boolean bool
					pod     []*corev1.Pod
				)
				BeforeEach(func() {
					boolean, err = r.checkUpdatedPods(pod, "0.7.1")
				})
				It("Error should be nil and bool value should be true", func() {
					Ω(err).Should(BeNil())
					Ω(boolean).Should(Equal(true))
				})
			})
			Context("getOneOutdatedPod", func() {
				var sts *appsv1.StatefulSet
				BeforeEach(func() {
					sts = &appsv1.StatefulSet{}
					r.client.Get(context.TODO(), types.NamespacedName{Name: util.StatefulSetNameForBookie(b.Name), Namespace: b.Namespace}, sts)
					_, err = r.getOneOutdatedPod(sts, "0.6.1")
				})
				It("Error should be nil", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("checkUpdatedPods with faulty pod", func() {
				var (
					boolean bool
					pod     []*corev1.Pod
				)
				BeforeEach(func() {
					testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Name: "test",
									State: v1.ContainerState{
										Waiting: &v1.ContainerStateWaiting{
											Reason: "CrashLoopBackOff",
										},
									},
								},
							}},
					}
					r.client.Create(context.TODO(), testpod)
					r.client.Get(context.TODO(), types.NamespacedName{Name: "test", Namespace: "default"}, testpod)
					pod = append(pod, testpod)
					boolean, err = r.checkUpdatedPods(pod, "0.7.1")
				})
				It("Error should not be nil and bool value should be false", func() {
					Ω(err).ShouldNot(BeNil())
					Ω(boolean).Should(Equal(false))
				})
			})
			Context("checkUpdatedPods with non-faulty pod", func() {
				var (
					boolean bool
					pod     []*corev1.Pod
				)
				BeforeEach(func() {
					testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Name: "test1",
								},
							},
						},
					}
					r.client.Create(context.TODO(), testpod)
					r.client.Get(context.TODO(), types.NamespacedName{Name: "test", Namespace: "default"}, testpod)
					pod = append(pod, testpod)
					boolean, err = r.checkUpdatedPods(pod, "0.7.1")
				})
				It("Error should not be nil and bool value should be false", func() {
					Ω(err).Should(BeNil())
					Ω(boolean).Should(Equal(false))
				})
			})
			Context("getStsPodsWithVersion", func() {
				var sts *appsv1.StatefulSet

				BeforeEach(func() {
					sts = &appsv1.StatefulSet{}
					r.client.Get(context.TODO(), types.NamespacedName{Name: util.StatefulSetNameForBookie(b.Name), Namespace: b.Namespace}, sts)
					_, err = r.getStsPodsWithVersion(sts, "0.6.1")
				})
				It("Error should be nil", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("syncBookkeeperVersion", func() {
				var (
					foundBookeeper        *v1alpha1.BookkeeperCluster
					err, err1, err2, err3 error
					b1, b2, b3, b4        bool
					sts                   *appsv1.StatefulSet
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					b1, err = r.syncBookkeeperVersion(b)
					foundBookeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookeeper)
					foundBookeeper.Status.TargetVersion = "0.7.1"
					r.client.Update(context.TODO(), foundBookeeper)
					b2, err1 = r.syncBookkeeperVersion(foundBookeeper)
					foundBookeeper.Status.TargetVersion = foundBookeeper.Spec.Version
					r.client.Update(context.TODO(), foundBookeeper)
					b3, err2 = r.syncBookkeeperVersion(foundBookeeper)
					sts = &appsv1.StatefulSet{}
					name := util.StatefulSetNameForBookie(foundBookeeper.Name)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: foundBookeeper.Namespace}, sts)
					sts.Status.UpdatedReplicas = sts.Status.Replicas - 1
					r.client.Update(context.TODO(), sts)
					b4, err3 = r.syncBookkeeperVersion(foundBookeeper)

				})
				It("Error should not be nil and b1 is false", func() {
					Ω(b1).Should(Equal(false))
					Ω(err.Error()).Should(Equal("target version is not set"))
				})
				It("Error should be nil and b2 is false", func() {
					Ω(b2).Should(Equal(false))
					Ω(err1).Should(BeNil())
				})
				It("Error should be nil and b3 is ture", func() {
					Ω(b3).Should(Equal(true))
					Ω(err2).Should(BeNil())
				})
				It("Error should not be nil and b4 is false", func() {
					Ω(b4).Should(Equal(false))
					Ω(err3).ShouldNot(BeNil())
				})
			})
		})
	})
})
