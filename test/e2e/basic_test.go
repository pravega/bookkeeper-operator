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

	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

// Test create and recreate a Bookkeeper cluster with the same name
var _ = Describe("Test create and recreate Bookkeeper cluster with the same name", func() {
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
			svcName    string
		)

		Context("When creating cluster first time", func() {
			It("should create successfully", func() {
				defaultCluster.Spec.HeadlessSvcNameSuffix = "headlesssvc"
				bookkeeper, err = e2eutil.CreateBKCluster(k8sClient, defaultCluster)
				Expect(err).NotTo(HaveOccurred())
				Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper), timeout).Should(Succeed())
			})

			It("should have the proper service", func() {
				svcName := fmt.Sprintf("%s-headlesssvc", bookkeeper.Name)
				err = e2eutil.CheckServiceExists(k8sClient, bookkeeper, svcName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should tear down the cluster successfully", func() {
				err = e2eutil.DeleteBKCluster(k8sClient, bookkeeper)
				Expect(err).NotTo(HaveOccurred())
				Eventually(e2eutil.WaitForBKClusterToTerminate(k8sClient, bookkeeper), timeout).Should(Succeed())
			})
		})

		Context("Recreating the cluster with same name", func() {
			It("Should create successfully", func() {
				bookkeeper, err = e2eutil.CreateBKCluster(k8sClient, defaultCluster)
				Expect(err).NotTo(HaveOccurred())
				Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper), timeout).Should(Succeed())
			})

			It("should have the proper service", func() {
				svcName = fmt.Sprintf("%s-bookie-headless", bookkeeper.Name)
				err = e2eutil.CheckServiceExists(k8sClient, bookkeeper, svcName)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should tear down the cluster successfully", func() {
				err = e2eutil.DeleteBKCluster(k8sClient, bookkeeper)
				Expect(err).NotTo(HaveOccurred())
				Eventually(e2eutil.WaitForBKClusterToTerminate(k8sClient, bookkeeper), timeout).Should(Succeed())
			})
		})
	})
})
