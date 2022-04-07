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

var _ = Describe("Delete pod test", func() {
	Context("Delete pod operations", func() {
		It("should delete pods ", func() {

			cluster := bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			cluster.WithDefaults()

			bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, cluster)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			podDeleteCount := 1
			err = bookkeeper_e2eutil.DeletePods(&t, k8sClient, bookkeeper, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(10 * time.Second)
			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			podDeleteCount = 2
			err = bookkeeper_e2eutil.DeletePods(&t, k8sClient, bookkeeper, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(10 * time.Second)

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			podDeleteCount = 3
			err = bookkeeper_e2eutil.DeletePods(&t, k8sClient, bookkeeper, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(10 * time.Second)

			err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
