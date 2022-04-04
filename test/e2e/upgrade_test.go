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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Upgrade  test controller", func() {
	Context("upgrade  operations", func() {
		It("upgrade pods  shoould be successful", func() {
			//By("create Zookeeper cluster")

			cluster := bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			cluster.WithDefaults()

			cluster.WithDefaults()
			initialVersion := "0.6.0"
			firstUpgradeVersion := "0.7.0"
			secondUpgradeVersion := "0.7.1"
			cluster.Spec.Version = initialVersion

			bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, cluster)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Status.CurrentVersion).To(Equal(initialVersion))

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			bookkeeper.Spec.Version = firstUpgradeVersion
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(2 * time.Second)

			// trigger another upgrade while this upgrade is happening- it should fail
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			bookkeeper.Spec.Version = secondUpgradeVersion
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).To(HaveOccurred(), "Should reject upgrade request while upgrade is in progress")
			Expect(err.Error()).To(ContainSubstring("failed to process the request, cluster is upgrading"))

			err = bookkeeper_e2eutil.WaitForBKClusterToUpgrade(&t, k8sClient, bookkeeper, firstUpgradeVersion)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			Expect(bookkeeper.Spec.Version).To(Equal(firstUpgradeVersion))
			Expect(bookkeeper.Status.CurrentVersion).To(Equal(firstUpgradeVersion))
			Expect(bookkeeper.Status.TargetVersion).To(Equal(""))

			// check version history
			Expect(bookkeeper.Status.VersionHistory[0]).To(Equal("0.6.0"))
			Expect(bookkeeper.Status.VersionHistory[1]).To(Equal("0.7.0"))

			// Delete cluster
			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
