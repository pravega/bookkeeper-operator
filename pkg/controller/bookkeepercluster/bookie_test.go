/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package bookkeepercluster_test

import (
	"testing"

	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/controller/bookkeepercluster"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBookkeeper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bookkeeper")
}

var _ = Describe("Bookie", func() {
	var _ = Describe("Bookie Test", func() {
		var (
			bk *v1alpha1.BookkeeperCluster
		)
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
		})
		Context("User is specifying bookkeeper journal and ledger path ", func() {

			var (
				customReq *corev1.ResourceRequirements
				err       error
			)
			BeforeEach(func() {
				customReq = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("6Gi"),
					},
				}
				boolFalse := false
				bk.Spec = v1alpha1.BookkeeperClusterSpec{
					Version:            "0.4.0",
					ServiceAccountName: "bk-operator",
					EnvVars:            "bk-configmap",
					AutoRecovery:       &boolFalse,
					Resources:          customReq,
					Options: map[string]string{
						"journalDirectories": "/bk/journal/j0,/bk/journal/j1,/bk/journal/j2,/bk/journal/j3",
						"ledgerDirectories":  "/bk/ledgers/l0,/bk/ledgers/l1,/bk/ledgers/l2,/bk/ledgers/l3",
					},
				}
				bk.WithDefaults()
			})
			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("Bookkeeper", func() {

				It("should create a headless service", func() {
					_ = bookkeepercluster.MakeBookieHeadlessService(bk)
					Ω(err).Should(BeNil())
				})

				It("should create a pod disruption budget", func() {
					_ = bookkeepercluster.MakeBookiePodDisruptionBudget(bk)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = bookkeepercluster.MakeBookieConfigMap(bk)
					Ω(err).Should(BeNil())
				})

				It("should create a stateful set", func() {
					_ = bookkeepercluster.MakeBookieStatefulSet(bk)
					Ω(err).Should(BeNil())
				})

			})
		})
		Context("User is not specifying bookkeeper journal and ledger path ", func() {

			var (
				err error
			)
			BeforeEach(func() {
				bk.Spec = v1alpha1.BookkeeperClusterSpec{}
				bk.WithDefaults()
			})
			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})
			Context("Bookkeeper", func() {

				It("should create a headless service", func() {
					_ = bookkeepercluster.MakeBookieHeadlessService(bk)
					Ω(err).Should(BeNil())
				})

				It("should create a pod disruption budget", func() {
					_ = bookkeepercluster.MakeBookiePodDisruptionBudget(bk)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = bookkeepercluster.MakeBookieConfigMap(bk)
					Ω(err).Should(BeNil())
				})

				It("should create a stateful set", func() {
					_ = bookkeepercluster.MakeBookieStatefulSet(bk)
					Ω(err).Should(BeNil())
				})

			})
		})
	})
})
