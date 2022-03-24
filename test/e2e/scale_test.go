/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Test scaling cluster", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)

	BeforeEach(func() {
		defaultCluster = e2eutil.NewDefaultCluster(namespace)
		defaultCluster.WithDefaults()
	})

	Context("Creating a bookkeeper cluster", func() {
		var (
			bookkeeper *v1alpha1.BookkeeperCluster
			err        error
		)

		It("should create successfully", func() {
			bookkeeper, err = e2eutil.CreateBKCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper), timeout).Should(Succeed())
		})

		It("should scale up successfully", func() {
			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// Scale up Bookkeeper cluster
			bookkeeper.Spec.Replicas = 5

			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper), timeout).Should(Succeed())
		})

		It("should scale down successfully", func() {
			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			// Scale down Bookkeeper cluster back to default
			bookkeeper.Spec.Replicas = 3

			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper), timeout).Should(Succeed())
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeleteBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForBKClusterToTerminate(k8sClient, bookkeeper), timeout).Should(Succeed())
		})
	})
})
