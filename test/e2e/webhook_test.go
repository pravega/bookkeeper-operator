/**
 * Copyright (c) 2019 Dell Inc., or its subsidiaries. All Rights Reserved.
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

var _ = Describe("Test deleted pods are auto restarted", func() {
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

		It("should fail to create", func() {
			// Test webhook with an invalid Bookkeeper cluster version format
			invalidVersion := e2eutil.NewClusterWithVersion(namespace, "999")
			invalidVersion.WithDefaults()
			_, err = e2eutil.CreateBKCluster(k8sClient, invalidVersion)
			Expect(err).To(HaveOccurred(), "Should reject deployment of invalid version format")
			Expect(err.Error()).To(ContainSubstring("request version is not in valid format:"))
		})

		It("should succeed with correct version", func() {
			// Test webhook with a valid Bookkeeper cluster version format
			validVersion := e2eutil.NewClusterWithVersion(namespace, "0.6.0")
			validVersion.WithDefaults()
			bookkeeper, err := e2eutil.CreateBKCluster(k8sClient, validVersion)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper), timeout).Should(Succeed())
		})

		It("should not allow downgrading", func() {
			// Try to downgrade the cluster
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			bookkeeper.Spec.Version = "0.5.0"
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).To(HaveOccurred(), "Should not allow downgrade")
			Expect(err.Error()).To(ContainSubstring("downgrading the cluster from version 0.6.0 to 0.5.0 is not supported"))
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeleteBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForBKClusterToTerminate(k8sClient, bookkeeper), timeout).Should(Succeed())
		})
	})
})
