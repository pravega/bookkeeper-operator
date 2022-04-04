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

	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Rollback controller", func() {
	Context("Check rollback operations", func() {
		It("should rollback without error", func() {

			cluster := bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			cluster.WithDefaults()

			bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, cluster)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// Scale up Bookkeeper cluster
			bookkeeper.Spec.Replicas = 5

			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// Scale down Bookkeeper cluster back to default
			bookkeeper.Spec.Replicas = 3

			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// Delete cluster
			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
