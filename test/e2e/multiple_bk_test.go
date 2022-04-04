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
	//	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"

	"strconv"
	"time"
)

var _ = Describe("Basic multiple BK controller", func() {
	Context("Check multiple BK cluster operations", func() {
		It("should create multiple clusters", func() {

			// Create first cluster
			cluster := bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			cm_name := "configmap1"
			cm1 := bookkeeper_e2eutil.NewConfigMap(testNamespace, cm_name, "pr1")
			err := bookkeeper_e2eutil.CreateConfigMap(&t, k8sClient, cm1)
			Expect(err).NotTo(HaveOccurred())
			cluster.ObjectMeta.Name = "bk1"
			autorecovery := true
			cluster.Spec.AutoRecovery = &(autorecovery)
			cluster.WithDefaults()

			bk1, err := bookkeeper_e2eutil.CreateBKClusterWithCM(&t, k8sClient, cluster, cm_name)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bk1)
			Expect(err).NotTo(HaveOccurred())

			bk1, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bk1)
			Expect(err).NotTo(HaveOccurred())
			err = bookkeeper_e2eutil.CheckConfigMap(&t, k8sClient, bk1, "BK_autoRecoveryDaemonEnabled", strconv.FormatBool(autorecovery))
			Expect(err).NotTo(HaveOccurred())

			// Create second cluster
			cluster = bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			cm_name = "configmap2"
			cm2 := bookkeeper_e2eutil.NewConfigMap(testNamespace, cm_name, "pr2")
			err = bookkeeper_e2eutil.CreateConfigMap(&t, k8sClient, cm2)
			Expect(err).NotTo(HaveOccurred())
			cluster.ObjectMeta.Name = "bk2"
			autorecovery = false
			cluster.Spec.AutoRecovery = &(autorecovery)
			cluster.WithDefaults()

			bk2, err := bookkeeper_e2eutil.CreateBKClusterWithCM(&t, k8sClient, cluster, cm_name)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bk2)
			Expect(err).NotTo(HaveOccurred())

			bk2, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bk2)
			Expect(err).NotTo(HaveOccurred())
			err = bookkeeper_e2eutil.CheckConfigMap(&t, k8sClient, bk2, "BK_autoRecoveryDaemonEnabled", strconv.FormatBool(autorecovery))
			Expect(err).NotTo(HaveOccurred())

			// Create third cluster
			cluster = bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			cluster.WithDefaults()

			bk3, err := bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, cluster)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bk3)
			Expect(err).NotTo(HaveOccurred())

			bk3, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bk3)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bk1, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bk1)
			Expect(err).NotTo(HaveOccurred())

			// Scale up replicas in the first Bookkeeper cluster
			bk1.Spec.Replicas = 5

			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bk1)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bk1)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bk2, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bk2)
			Expect(err).NotTo(HaveOccurred())

			// Deleting pods of the second Bookkeeper cluster
			podDeleteCount := 3
			err = bookkeeper_e2eutil.DeletePods(&t, k8sClient, bk2, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(10 * time.Second)

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bk2)
			Expect(err).NotTo(HaveOccurred())

			// deleting all bookkeeper clusters
			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bk1)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bk1)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bk2)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bk2)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bk3)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bk3)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.DeleteConfigMap(&t, k8sClient, cm1)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.DeleteConfigMap(&t, k8sClient, cm2)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
