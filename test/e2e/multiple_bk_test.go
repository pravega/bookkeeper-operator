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
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"

	"github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Test multiple bookkeeper clusters", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)

	BeforeEach(func() {
		defaultCluster = e2eutil.NewDefaultCluster(namespace)
		defaultCluster.WithDefaults()
	})

	Context("When creating 3 Bookkeeper clusters", func() {
		var (
			bk1, bk2, bk3 *v1alpha1.BookkeeperCluster
			err           error
		)
		cm1_name := "configmap1"
		cm1 := e2eutil.NewConfigMap(namespace, cm1_name, "pr1")
		cm2_name := "configmap2"
		cm2 := e2eutil.NewConfigMap(namespace, cm2_name, "pr2")

		Context("When creating cluster 1", func() {
			autorecovery := true
			It("should succeed", func() {
				err = e2eutil.CreateConfigMap(k8sClient, cm1)
				Expect(err).NotTo(HaveOccurred())
				defaultCluster.ObjectMeta.Name = "bk1"
				defaultCluster.Spec.AutoRecovery = &(autorecovery)
				defaultCluster.WithDefaults()

				bk1, err = e2eutil.CreateBKClusterWithCM(k8sClient, defaultCluster, cm1_name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bk1), timeout).Should(Succeed())
			})
			It("should have the proper configmap", func() {
				bk1, err = e2eutil.GetBKCluster(k8sClient, bk1)
				Expect(err).NotTo(HaveOccurred())
				err = e2eutil.CheckConfigMap(k8sClient, bk1, "BK_autoRecoveryDaemonEnabled", strconv.FormatBool(autorecovery))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When creating cluster 2", func() {
			autorecovery := false
			It("should succeed", func() {
				defaultCluster = e2eutil.NewDefaultCluster(namespace)
				err = e2eutil.CreateConfigMap(k8sClient, cm2)
				Expect(err).NotTo(HaveOccurred())
				defaultCluster.ObjectMeta.Name = "bk2"
				defaultCluster.Spec.AutoRecovery = &(autorecovery)
				defaultCluster.WithDefaults()

				bk2, err := e2eutil.CreateBKClusterWithCM(k8sClient, defaultCluster, cm2_name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bk2), timeout).Should(Succeed())
			})
			It("should have the proper configmap", func() {
				bk2, err = e2eutil.GetBKCluster(k8sClient, bk2)
				Expect(err).NotTo(HaveOccurred())
				err = e2eutil.CheckConfigMap(k8sClient, bk2, "BK_autoRecoveryDaemonEnabled", strconv.FormatBool(autorecovery))
				Expect(err).NotTo(HaveOccurred())
			})
			Context("When creating cluster 3", func() {
				It("should succeed", func() {
					defaultCluster = e2eutil.NewDefaultCluster(namespace)
					defaultCluster.WithDefaults()

					bk3, err := e2eutil.CreateBKCluster(k8sClient, defaultCluster)
					Expect(err).NotTo(HaveOccurred())
					Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bk3), timeout).Should(Succeed())
				})

				It("should update & modify each cluster successfully", func() {
					bk3, err = e2eutil.GetBKCluster(k8sClient, bk3)
					Expect(err).NotTo(HaveOccurred())

					// This is to get the latest Bookkeeper cluster object
					bk1, err = e2eutil.GetBKCluster(k8sClient, bk1)
					Expect(err).NotTo(HaveOccurred())

					// Scale up replicas in the first Bookkeeper cluster
					bk1.Spec.Replicas = 5

					err = e2eutil.UpdateBKCluster(k8sClient, bk1)
					Expect(err).NotTo(HaveOccurred())

					Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bk1), timeout).Should(Succeed())

					// This is to get the latest Bookkeeper cluster object
					bk2, err = e2eutil.GetBKCluster(k8sClient, bk2)
					Expect(err).NotTo(HaveOccurred())

					// Deleting pods of the second Bookkeeper cluster
					podDeleteCount := 3
					err = e2eutil.DeletePods(k8sClient, bk2, podDeleteCount)
					Expect(err).NotTo(HaveOccurred())
					time.Sleep(10 * time.Second)

					Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bk2), timeout).Should(Succeed())
				})

				It("should tear down all the clusters", func() {
					// deleting all bookkeeper clusters
					err = e2eutil.DeleteBKCluster(k8sClient, bk1)
					Expect(err).NotTo(HaveOccurred())

					err = e2eutil.WaitForBKClusterToTerminate(k8sClient, bk1)
					Expect(err).NotTo(HaveOccurred())

					err = e2eutil.DeleteBKCluster(k8sClient, bk2)
					Expect(err).NotTo(HaveOccurred())

					err = e2eutil.WaitForBKClusterToTerminate(k8sClient, bk2)
					Expect(err).NotTo(HaveOccurred())

					err = e2eutil.DeleteBKCluster(k8sClient, bk3)
					Expect(err).NotTo(HaveOccurred())

					err = e2eutil.WaitForBKClusterToTerminate(k8sClient, bk3)
					Expect(err).NotTo(HaveOccurred())

					err = e2eutil.DeleteConfigMap(k8sClient, cm1)
					Expect(err).NotTo(HaveOccurred())

					err = e2eutil.DeleteConfigMap(k8sClient, cm2)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
