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
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

// Test create and recreate a Bookkeeper cluster with the same name

var _ = Describe("Test create and recreate Bookkeeper cluster with the same name", func() {
	Context("Check create/delete operations", func() {
		It("should create and delete operations should be successful", func() {
			By("create Bookkeeper cluster")
			defaultCluster := bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()
			defaultCluster.Spec.HeadlessSvcNameSuffix = "headlesssvc"

			bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Expect(bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, defaultCluster)).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			svcName := fmt.Sprintf("%s-headlesssvc", bookkeeper.Name)
			err = bookkeeper_e2eutil.CheckServiceExists(&t, k8sClient, bookkeeper, svcName)
			Expect(err).NotTo(HaveOccurred())
      By("delete created Bookkeeper cluster")
			Expect(k8sClient.Delete(ctx, bookkeeper)).Should(Succeed())
			Expect(bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bookkeeper)).NotTo(HaveOccurred())

			By("create Bookkeeper cluster with the same name")
			defaultCluster = bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()

			bookkeeper, err = bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, defaultCluster)).NotTo(HaveOccurred())
			svcName = fmt.Sprintf("%s-bookie-headless", bookkeeper.Name)
			err = bookkeeper_e2eutil.CheckServiceExists(&t, k8sClient, bookkeeper, svcName)
			Expect(err).NotTo(HaveOccurred())
			By("delete created Bookkeeper cluster")
			Expect(k8sClient.Delete(ctx, bookkeeper)).Should(Succeed())
			Expect(bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bookkeeper)).NotTo(HaveOccurred())
		})
	})
})
