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
	corev1 "k8s.io/api/core/v1"

	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Test Bookkeeper cluster rollback", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)

	BeforeEach(func() {
		defaultCluster = e2eutil.NewDefaultCluster(namespace)
		defaultCluster.WithDefaults()
	})
	Context("Create a bookkeeper cluster", func() {

		var (
			bookkeeper *v1alpha1.BookkeeperCluster
			err        error
		)
		initialVersion := "0.6.0"
		firstUpgradeVersion := "0.7.0-1"
		secondUpgradeVersion := "0.5.0"
		It("should create a cluster", func() {
			defaultCluster.Spec.Version = initialVersion
			bookkeeper, err := e2eutil.CreateBKCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper), timeout).Should(Succeed())
		})

		It("should have the right version", func() {
			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Status.CurrentVersion).To(Equal(initialVersion))

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail an upgarde", func() {
			bookkeeper.Spec.Version = firstUpgradeVersion
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// waiting for upgrade to fail
			time.Sleep(2 * time.Minute)

			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			_, errorCondition := bookkeeper.Status.GetClusterCondition(v1alpha1.ClusterConditionError)
			Expect(errorCondition.Status).To(Equal(corev1.ConditionTrue))
			Expect(errorCondition.Reason).To(Equal("UpgradeFailed"))
			Expect(errorCondition.Message).To(ContainSubstring("pod bookkeeper-bookie-0 update failed because of ImagePullBackOff"))
		})

		It("should have sent an upgrade error event", func() {
			// checking whether upgrade error event is sent out to the kubernetes event queue
			event, err := e2eutil.CheckEvents(k8sClient, bookkeeper, "UPGRADE_ERROR")
			Expect(err).NotTo(HaveOccurred())
			Expect(event).To(BeTrue())

			// trigger rollback to version other than last stable version
			bookkeeper.Spec.Version = secondUpgradeVersion
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).To(HaveOccurred(), "Should not allow rollback to any version other than the last stable version")
			Expect(err.Error()).To(ContainSubstring("Rollback to version 0.5.0 not supported. Only rollback to version 0.6.0 is supported"))
		})

		It("should trigger a rollback to the last stable version", func() {
			// trigger rollback to last stable version
			bookkeeper.Spec.Version = initialVersion
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(2 * time.Second)

			// trigger another upgrade while the last rollback is still ongoing
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			bookkeeper.Spec.Version = secondUpgradeVersion
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).To(HaveOccurred(), "Should reject rollback request while rollback is in progress")
			Expect(err.Error()).To(ContainSubstring("failed to process the request, rollback in progress"))

			_, rollbackCondition := bookkeeper.Status.GetClusterCondition(v1alpha1.ClusterConditionRollback)
			Expect(rollbackCondition.Status).To(Equal(corev1.ConditionTrue))
			Expect(rollbackCondition.Reason).To(ContainSubstring("Updating Bookkeeper"))

			Eventually(e2eutil.WaitForBKClusterToRollback(k8sClient, bookkeeper, initialVersion), timeout).Should(Succeed())

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// waiting for rollback to complete
			Expect(bookkeeper.Spec.Version).To(Equal(initialVersion))
			Expect(bookkeeper.Status.CurrentVersion).To(Equal(initialVersion))
			Expect(bookkeeper.Status.TargetVersion).To(Equal(""))

			// checking version history
			Expect(bookkeeper.Status.VersionHistory[0]).To(Equal("0.6.0"))
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeleteBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForBKClusterToTerminate(k8sClient, bookkeeper), timeout).Should(Succeed())
		})
	})
})
