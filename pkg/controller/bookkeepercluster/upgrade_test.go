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
	"testing"

	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/util"

	bookkeeperv1alpha1 "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			p   *v1alpha1.BookkeeperCluster
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p.Spec.Version = "0.5.0"
			s.AddKnownTypes(v1alpha1.SchemeGroupVersion, p)
		})

		Context("Cluster condition prior to Upgrade", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
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
					foundBookkeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, err = r.Reconcile(req)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should have current version set to spec version", func() {
					Ω(foundBookkeeper.Status.CurrentVersion).Should(Equal(foundBookkeeper.Spec.Version))
				})

				It("should set upgrade condition and status to be false", func() {
					_, upgradeCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Status).Should(Equal(corev1.ConditionFalse))
				})
			})
		})

		Context("Upgrade to new version", func() {
			var (
				client client.Client
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.BookkeeperClusterSpec{
					Version: "0.5.0",
				}
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcileBookkeeperCluster{client: client, scheme: s}
				_, _ = r.Reconcile(req)
				foundBookkeeper := &v1alpha1.BookkeeperCluster{}
				_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				foundBookkeeper.Spec.Version = "0.6.0"
				// bypass the pods ready check in the upgrade logic
				foundBookkeeper.Status.SetPodsReadyConditionTrue()
				client.Update(context.TODO(), foundBookkeeper)
				_, _ = r.Reconcile(req)
			})

			Context("Upgrading Condition", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set target version to 0.6.0", func() {
					Ω(foundBookkeeper.Status.TargetVersion).Should(Equal("0.6.0"))
				})

				It("should set upgrade condition to be true", func() {
					_, upgradeCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Status).Should(Equal(corev1.ConditionTrue))
				})
			})

			Context("Upgrade Bookkeeper", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
					sts             *appsv1.StatefulSet
				)
				BeforeEach(func() {
					sts = &appsv1.StatefulSet{}
					name := util.StatefulSetNameForBookie(p.Name)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					sts.Status.ReadyReplicas = 1
					r.client.Update(context.TODO(), sts)

					_, _ = r.Reconcile(req)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set upgrade condition reason to UpgradingBookkeeperReason and message to 0", func() {
					_, upgradeCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(bookkeeperv1alpha1.UpdatingBookkeeperReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})

			Context("Upgrade Segmentstore", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
					sts             *appsv1.StatefulSet
				)
				BeforeEach(func() {
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)

					// Bookkeeper
					sts = &appsv1.StatefulSet{}
					name := util.StatefulSetNameForBookie(p.Name)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					targetImage, _ := util.BookkeeperTargetImage(foundBookkeeper)
					sts.Spec.Template.Spec.Containers[0].Image = targetImage
					r.client.Update(context.TODO(), sts)

					_, _ = r.Reconcile(req)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set upgrade condition reason to UpgradingSegmentstoreReason and message to 0", func() {
					_, upgradeCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(bookkeeperv1alpha1.UpdatingSegmentstoreReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})

			Context("Upgrade Controller", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
					sts             *appsv1.StatefulSet
				)
				BeforeEach(func() {
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)

					// Bookkeeper
					sts = &appsv1.StatefulSet{}
					name := util.StatefulSetNameForBookie(p.Name)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					targetImage, _ := util.BookkeeperTargetImage(foundBookkeeper)
					sts.Spec.Template.Spec.Containers[0].Image = targetImage
					r.client.Update(context.TODO(), sts)

					_, _ = r.Reconcile(req)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set upgrade condition reason to UpgradingControllerReason and message to 0", func() {
					_, upgradeCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(bookkeeperv1alpha1.UpdatingControllerReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})
		})
	})

	var _ = Describe("Rollback Test", func() {
		var (
			req reconcile.Request
			p   *v1alpha1.BookkeeperCluster
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p.Spec.Version = "0.5.0"
			s.AddKnownTypes(v1alpha1.SchemeGroupVersion, p)
		})

		Context("Cluster Condition before Rollback", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
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
					foundBookkeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, err = r.Reconcile(req)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should have current version set to spec version", func() {
					Ω(foundBookkeeper.Status.CurrentVersion).Should(Equal(foundBookkeeper.Spec.Version))
				})

				It("should not have rollback condition set", func() {
					_, rollbackCondition := foundBookkeeper.Status.GetClusterCondition(v1alpha1.ClusterConditionRollback)
					Ω(rollbackCondition).Should(BeNil())
				})

				It("should have version history set", func() {
					history := foundBookkeeper.Status.VersionHistory
					Ω(history[0]).Should(Equal("0.5.0"))
				})

			})
		})

		Context("Rollback to previous version", func() {
			var (
				client client.Client
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.BookkeeperClusterSpec{
					Version: "0.6.0",
				}
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcileBookkeeperCluster{client: client, scheme: s}
				_, _ = r.Reconcile(req)
				foundBookkeeper := &v1alpha1.BookkeeperCluster{}
				_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				foundBookkeeper.Spec.Version = "0.5.0"
				foundBookkeeper.Status.VersionHistory = []string{"0.5.0"}
				// bypass the pods ready check in the upgrade logic
				foundBookkeeper.Status.SetPodsReadyConditionFalse()
				foundBookkeeper.Status.SetErrorConditionTrue("UpgradeFailed", "some error")
				client.Update(context.TODO(), foundBookkeeper)
				_, _ = r.Reconcile(req)
			})

			Context("Rollback Triggered", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set Rollback condition status to be true", func() {
					_, rollbackCondition := foundBookkeeper.Status.GetClusterCondition(v1alpha1.ClusterConditionRollback)
					Ω(rollbackCondition.Status).To(Equal(corev1.ConditionTrue))
				})

				It("should set target version to previous version", func() {
					Ω(foundBookkeeper.Status.TargetVersion).To(Equal(foundBookkeeper.Spec.Version))
				})
			})

			Context("Rollback Controller", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set rollback condition reason to UpdatingController and message to 0", func() {
					_, rollbackCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionRollback)
					Ω(rollbackCondition.Reason).Should(Equal(bookkeeperv1alpha1.UpdatingControllerReason))
					Ω(rollbackCondition.Message).Should(Equal("0"))
				})
			})

			Context("Rollback SegmentStore", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set rollback condition reason to UpdatingSegmentStore and message to 0", func() {
					_, rollbackCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionRollback)
					Ω(rollbackCondition.Reason).Should(Equal(bookkeeperv1alpha1.UpdatingSegmentstoreReason))
					Ω(rollbackCondition.Message).Should(Equal("0"))
				})
			})

			Context("Rollback Bookkeeper", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set rollback condition reason to UpdatingBookkeeper and message to 0", func() {
					_, rollbackCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionRollback)
					Ω(rollbackCondition.Reason).Should(Equal(bookkeeperv1alpha1.UpdatingBookkeeperReason))
					Ω(rollbackCondition.Message).Should(Equal("0"))
				})
			})
			Context("Rollback Completed", func() {
				var (
					foundBookkeeper *v1alpha1.BookkeeperCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})

				It("should set currentversion equal to target version", func() {
					Ω(foundBookkeeper.Status.CurrentVersion).Should(Equal("0.5.0"))
				})
				It("should set TargetVersoin to empty", func() {
					Ω(foundBookkeeper.Status.TargetVersion).Should(Equal(""))
				})
				It("should set rollback condition to false", func() {
					_, rollbackCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionRollback)
					Ω(rollbackCondition.Status).To(Equal(corev1.ConditionFalse))
				})
				It("should set error condition to false", func() {
					_, errorCondition := foundBookkeeper.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionError)
					Ω(errorCondition.Status).To(Equal(corev1.ConditionFalse))
				})
			})
		})
	})
})
