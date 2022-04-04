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
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("webhook  test controller", func() {
	Context("webhook validation operations", func() {
		It("should throw proper error message with invalid config", func() {
			By("create Zookeeper cluster")
			cluster := bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			cluster.WithDefaults()

			//Test webhook with an invalid Bookkeeper cluster version format
			invalidVersion := bookkeeper_e2eutil.NewClusterWithVersion(testNamespace, "999")
			invalidVersion.WithDefaults()
			_, err := bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, invalidVersion)
			Expect(err).To(HaveOccurred(), "Should reject deployment of invalid version format")
			Expect(err.Error()).To(ContainSubstring("request version is not in valid format:"))

			// Test webhook with a valid Bookkeeper cluster version format
			validVersion := bookkeeper_e2eutil.NewClusterWithVersion(testNamespace, "0.6.0")
			validVersion.WithDefaults()
			bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, validVersion)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// Try to downgrade the cluster
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			bookkeeper.Spec.Version = "0.5.0"
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).To(HaveOccurred(), "Should not allow downgrade")
			Expect(err.Error()).To(ContainSubstring("downgrading the cluster from version 0.6.0 to 0.5.0 is not supported"))

			// Delete cluster
			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
