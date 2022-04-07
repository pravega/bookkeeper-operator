/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/pravega/bookkeeper-operator/pkg/controller/config"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/bookkeeper-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("BookkeeperCluster Controller", func() {
	const (
		Name      = "example"
		Namespace = "default"
	)
	var (
		s = scheme.Scheme
		r *BookkeeperClusterReconciler
	)

	Context("Reconcile", func() {
		var (
			req reconcile.Request
			res reconcile.Result
			b   *v1alpha1.BookkeeperCluster
		//	ctx context.Context
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			b = &v1alpha1.BookkeeperCluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "BookkeeperCluster",
					APIVersion: "bookkeeper.pravega.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			s.AddKnownTypes(v1alpha1.GroupVersion, b)
		})
		Context("Without spec", func() {
			var (
				client          client.Client
				err             error
				foundBookkeeper *v1alpha1.BookkeeperCluster
				ctx             context.Context
			)

			BeforeEach(func() {
				client = fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(b).Build()
				//client = fake.NewFakeClient(b)

				r = &BookkeeperClusterReconciler{Client: client, Scheme: s}
				//1st reconcile
				res, err = r.Reconcile(ctx, req)
			})
			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})

			Context("Before defaults are applied", func() {
				It("should requeue the request", func() {
					Ω(res.Requeue).Should(BeTrue())
				})

				It("should set the default cluster spec options", func() {
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					err = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
					Ω(err).Should(BeNil())
					Ω(foundBookkeeper.Spec).ShouldNot(BeNil())
					Ω(foundBookkeeper.Spec.Version).Should(Equal(v1alpha1.DefaultBookkeeperVersion))
					Ω(foundBookkeeper.Spec.ZookeeperUri).Should(Equal(v1alpha1.DefaultZookeeperUri))
				})
			})
			Context("After defaults are applied", func() {
				BeforeEach(func() {
					// 2nd reconcile
					res, err = r.Reconcile(ctx, req)
				})
				It("should requeue after ReconfileTime delay", func() {
					Ω(res.RequeueAfter).To(Equal(ReconcileTime))
				})
				It("should set current version on 2nd reconcile ", func() {
					res, err = r.Reconcile(ctx, req)
					foundBookkeeper := &v1alpha1.BookkeeperCluster{}
					err = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
					Ω(err).Should(BeNil())
					Ω(foundBookkeeper.Status.CurrentVersion).Should(Equal(v1alpha1.DefaultBookkeeperVersion))
				})
			})
			Context("Checking Cluster deployment", func() {
				BeforeEach(func() {
					res, err = r.Reconcile(ctx, req)
					foundBookkeeper = &v1alpha1.BookkeeperCluster{}
					err = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
				})
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("checking updatePDB", func() {
				var (
					err1 error
					str1 string
				)
				BeforeEach(func() {
					res, err = r.Reconcile(ctx, req)
					currentpdb := &policyv1beta1.PodDisruptionBudget{}
					pdbname := fmt.Sprintf("%s-bookie", b.Name)
					r.Client.Get(context.TODO(), types.NamespacedName{Name: pdbname, Namespace: b.Namespace}, currentpdb)
					maxUnavailable := intstr.FromInt(3)
					newpdb := &policyv1beta1.PodDisruptionBudget{
						TypeMeta: metav1.TypeMeta{
							Kind:       "PodDisruptionBudget",
							APIVersion: "policy/v1beta1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-name",
							Namespace: b.Namespace,
						},
						Spec: policyv1beta1.PodDisruptionBudgetSpec{
							MaxUnavailable: &maxUnavailable,
							Selector: &metav1.LabelSelector{
								MatchLabels: b.LabelsForBookie(),
							},
						},
					}
					err1 = r.updatePdb(currentpdb, newpdb)
					str1 = fmt.Sprintf("%s", currentpdb.Spec.MaxUnavailable)
				})
				It("should not give error", func() {
					Ω(err1).Should(BeNil())
				})
				It("unavailable replicas should change to 3", func() {
					Ω(str1).To(Equal("3"))
				})
			})
			Context("checking checkVersionUpgradeTriggered function", func() {
				var (
					ans1, ans2 bool
				)
				BeforeEach(func() {
					ans1 = r.checkVersionUpgradeTriggered(b)
					b.Spec.Version = "0.8.0"
					ans2 = r.checkVersionUpgradeTriggered(b)
				})
				It("ans1 should be false", func() {
					Ω(ans1).To(Equal(false))
				})
				It("ans2 should be true", func() {
					Ω(ans2).To(Equal(true))
				})

			})

			Context("syncBookieSize", func() {
				var (
					err1 error
					flag bool
				)
				BeforeEach(func() {
					flag = false
					b.Spec.BlockOwnerDeletion = &flag
					b.WithDefaults()
					//to ensure the client get for BookKeepercluster fails
					err = r.syncBookieSize(b)
					_, _ = r.Reconcile(ctx, req)
					b.Spec.Replicas = 5
					client.Update(context.TODO(), b)
					_, err1 = r.Reconcile(ctx, req)
				})

				It("should give error", func() {
					Ω(strings.ContainsAny(err.Error(), "failed to get stateful-set")).Should(Equal(true))
				})
				It("should not give error", func() {
					Ω(err1).Should(BeNil())
				})
			})
			Context("reconcileFinalizers", func() {
				BeforeEach(func() {
					b.WithDefaults()
					config.DisableFinalizer = false
					b.Spec.EnvVars = "vars"
					client.Update(context.TODO(), b)
					_, err = r.Reconcile(ctx, req)
					now := metav1.Now()
					b.SetDeletionTimestamp(&now)
					client.Update(context.TODO(), b)
					_, err = r.Reconcile(ctx, req)
				})
				It("should not give error", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("reconcileFinalizers", func() {
				BeforeEach(func() {
					b.WithDefaults()
				})
				It("should have 1 finalizer", func() {
					config.DisableFinalizer = false
					err = r.reconcileFinalizers(b)
					Expect(b.ObjectMeta.Finalizers).To(HaveLen(1))
					Expect(err).NotTo(HaveOccurred())
				})
				It("should have 0 finalizer", func() {
					config.DisableFinalizer = true
					err = r.reconcileFinalizers(b)
					Expect(b.ObjectMeta.Finalizers).To(HaveLen(0))
					Expect(err).NotTo(HaveOccurred())
				})
			})
			Context("syncStatefulSetExternalServices withthout external service", func() {
				BeforeEach(func() {
					b.WithDefaults()
					s := MakeBookieStatefulSet(b)
					err = r.syncStatefulSetExternalServices(s)
				})
				It("should not give error", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("syncStatefulSetExternalServices with external service", func() {
				var sts *appsv1.StatefulSet
				BeforeEach(func() {
					svcDelete := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "svc-4",
							Namespace: b.Namespace,
						},
					}
					r.Client.Create(context.TODO(), svcDelete)
					b.WithDefaults()
					sts = &appsv1.StatefulSet{}
					sts = MakeBookieStatefulSet(b)
					r.Client.Create(context.TODO(), sts)
					name := b.Name
					_ = r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: b.Namespace}, sts)
					err = r.syncStatefulSetExternalServices(sts)
				})
				It("should not give error", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("syncStatefulSetPvc", func() {
				var sts *appsv1.StatefulSet
				BeforeEach(func() {
					pvc := &corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc-4",
							Namespace: b.Namespace,
						},
					}
					r.Client.Create(context.TODO(), pvc)
					b.WithDefaults()
					sts = &appsv1.StatefulSet{}
					sts = MakeBookieStatefulSet(b)
					r.Client.Create(context.TODO(), sts)
					name := b.Name
					_ = r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: b.Namespace}, sts)
					err = r.syncStatefulSetPvc(sts)
				})
				It("should not give error", func() {
					Ω(err).Should(BeNil())
				})
			})

			Context("rollbackFailedUpgrade", func() {
				BeforeEach(func() {
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
					foundBookkeeper.Status.CurrentVersion = "0.9.0"
					foundBookkeeper.Status.Init()
					foundBookkeeper.Status.AddToVersionHistory("0.6.0")
					foundBookkeeper.Spec.Version = foundBookkeeper.Status.GetLastVersion()
					foundBookkeeper.Status.SetErrorConditionTrue("UpgradeFailed", " ")
					r.Client.Update(context.TODO(), foundBookkeeper)
					err = r.rollbackFailedUpgrade(foundBookkeeper)
				})
				It("should not give error", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("getFinalizerAndClusterName", func() {
				var str1, str2 string
				BeforeEach(func() {
					b.WithDefaults()
					config.DisableFinalizer = false
					client.Update(context.TODO(), b)
					r.reconcileFinalizers(b)
					_ = client.Get(context.TODO(), req.NamespacedName, foundBookkeeper)
					str1, str2 = getFinalizerAndClusterName(foundBookkeeper.Finalizers)
				})
				It("should have str1 as cleanUpZookeeper ", func() {
					Ω(str1).Should(Equal("cleanUpZookeeper"))
				})
				It("should have str2 as pravega-cluster ", func() {
					Ω(str2).Should(Equal("pravega-cluster"))
				})
			})
			Context("Should have Reconcile Result false when request namespace does not contain bk cluster", func() {
				BeforeEach(func() {
					client = fake.NewFakeClient(b)
					r = &BookkeeperClusterReconciler{Client: client, Scheme: s}
					req.NamespacedName.Namespace = "temp"
					res, err = r.Reconcile(ctx, req)
				})
				It("should have false in reconcile result", func() {
					Ω(res.Requeue).To(Equal(false))
					Ω(err).To(BeNil())
				})
			})
			Context("reconcileFinalizers", func() {
				BeforeEach(func() {
					b.WithDefaults()
					config.DisableFinalizer = false
					client.Update(context.TODO(), b)
					err = r.reconcileFinalizers(b)
					now := metav1.Now()
					b.SetDeletionTimestamp(&now)
					client.Update(context.TODO(), b)
					err = r.reconcileFinalizers(b)

				})
				It("should give error due to failure in connecting to zookeeper", func() {
					Expect(err).To(HaveOccurred())
				})
			})
			Context("cleanUpZookeeperMeta", func() {
				BeforeEach(func() {
					b.WithDefaults()
					err = r.cleanUpZookeeperMeta(b, "pravega")
				})
				It("should give error", func() {
					Ω(err).ShouldNot(BeNil())
				})
			})
		})
	})
})
