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

	bkapi "github.com/pravega/bookkeeper-operator/api/v1alpha1"
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Rollback controller", func() {
	Context("Check rollback operations", func() {
		It("should rollback without error", func() {

			cluster := bookkeeper_e2eutil.NewDefaultCluster(testNamespace)

			cluster.WithDefaults()
			initialVersion := "0.6.0"
			firstUpgradeVersion := "0.7.0-1"
			secondUpgradeVersion := "0.5.0"
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

			// waiting for upgrade to fail
			time.Sleep(2 * time.Minute)

			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			_, errorCondition := bookkeeper.Status.GetClusterCondition(bkapi.ClusterConditionError)
			Expect(errorCondition.Status).To(Equal(corev1.ConditionTrue))
			Expect(errorCondition.Reason).To(Equal("UpgradeFailed"))
			Expect(errorCondition.Message).To(ContainSubstring("pod bookkeeper-bookie-0 update failed because of ImagePullBackOff"))

			// checking whether upgrade error event is sent out to the kubernetes event queue
			/*	event, err := bookkeeper_e2eutil.CheckEvents(&t, k8sClient, bookkeeper, "UPGRADE_ERROR")
				Expect(err).NotTo(HaveOccurred())
				Expect(event).To(BeTrue())*/

			// trigger rollback to version other than last stable version
			bookkeeper.Spec.Version = secondUpgradeVersion
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).To(HaveOccurred(), "Should not allow rollback to any version other than the last stable version")
			Expect(err.Error()).To(ContainSubstring("Rollback to version 0.5.0 not supported. Only rollback to version 0.6.0 is supported"))

			// trigger rollback to last stable version
			bookkeeper.Spec.Version = initialVersion
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(2 * time.Second)

			// trigger another upgrade while the last rollback is still ongoing
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			bookkeeper.Spec.Version = secondUpgradeVersion
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).To(HaveOccurred(), "Should reject rollback request while rollback is in progress")
			Expect(err.Error()).To(ContainSubstring("failed to process the request, rollback in progress"))

			_, rollbackCondition := bookkeeper.Status.GetClusterCondition(bkapi.ClusterConditionRollback)
			Expect(rollbackCondition.Status).To(Equal(corev1.ConditionTrue))
			Expect(rollbackCondition.Reason).To(ContainSubstring("Updating Bookkeeper"))

			err = bookkeeper_e2eutil.WaitForBKClusterToRollback(&t, k8sClient, bookkeeper, initialVersion)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// waiting for rollback to complete
			Expect(bookkeeper.Spec.Version).To(Equal(initialVersion))
			Expect(bookkeeper.Status.CurrentVersion).To(Equal(initialVersion))
			Expect(bookkeeper.Status.TargetVersion).To(Equal(""))

			// checking version history
			Expect(bookkeeper.Status.VersionHistory[0]).To(Equal("0.6.0"))

			// Delete cluster
			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
