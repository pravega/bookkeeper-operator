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
		initialVersion := "0.6.0"
		firstUpgradeVersion := "0.7.0"
		secondUpgradeVersion := "0.7.1"

		It("should create successfully", func() {
			defaultCluster.WithDefaults()

			defaultCluster.Spec.Version = initialVersion

			bookkeeper, err = e2eutil.CreateBKCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should upgrade successfully", func() {
			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Status.CurrentVersion).To(Equal(initialVersion))

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			bookkeeper.Spec.Version = firstUpgradeVersion
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(2 * time.Second)
		})

		It("should fail if another upgrade is started while upgrading", func() {
			// trigger another upgrade while this upgrade is happening- it should fail
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			bookkeeper.Spec.Version = secondUpgradeVersion
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).To(HaveOccurred(), "Should reject upgrade request while upgrade is in progress")
			Expect(err.Error()).To(ContainSubstring("failed to process the request, cluster is upgrading"))

			err = e2eutil.WaitForBKClusterToUpgrade(k8sClient, bookkeeper, firstUpgradeVersion)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should finish upgrading successfully", func() {
			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			Expect(bookkeeper.Spec.Version).To(Equal(firstUpgradeVersion))
			Expect(bookkeeper.Status.CurrentVersion).To(Equal(firstUpgradeVersion))
			Expect(bookkeeper.Status.TargetVersion).To(Equal(""))

			// check version history
			Expect(bookkeeper.Status.VersionHistory[0]).To(Equal("0.6.0"))
			Expect(bookkeeper.Status.VersionHistory[1]).To(Equal("0.7.0"))
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeleteBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForBKClusterToTerminate(k8sClient, bookkeeper), timeout).Should(Succeed())
		})
	})
})
