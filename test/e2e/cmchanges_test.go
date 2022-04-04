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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Conigmap  test controller", func() {
	Context("Check configmap update operations", func() {
		It("should create and recreate a Zookeeper cluster with the same name", func() {
			By("create Zookeeper cluster")
			cluster := bookkeeper_e2eutil.NewDefaultCluster(testNamespace)
			cluster.WithDefaults()
			initialVersion := "0.6.0"
			upgradeVersion := "0.7.0"
			gcOpts := []string{"-XX:+UseG1GC", "-XX:MaxGCPauseMillis=10"}
			gcOptions := strings.Join(gcOpts, " ")
			cluster.Spec.Version = initialVersion
			cluster.Spec.Options["minorCompactionThreshold"] = "0.4"
			cluster.Spec.Options["journalDirectories"] = "/bk/journal"
			cluster.Spec.Options["useHostNameAsBookieID"] = "true"
			cluster.Spec.JVMOptions.GcOpts = gcOpts

			bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(&t, k8sClient, cluster)
			Expect(err).NotTo(HaveOccurred())

			Expect(bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(&t, k8sClient, cluster)).NotTo(HaveOccurred())

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.CheckConfigMap(&t, k8sClient, bookkeeper, "BOOKIE_GC_OPTS", gcOptions)
			Expect(err).NotTo(HaveOccurred())

			// updating modifiable bookkeeper option
			gcOpts = []string{"-XX:-UseParallelGC", "-XX:MaxGCPauseMillis=10"}
			gcOptions = strings.Join(gcOpts, " ")
			bookkeeper.Spec.Version = upgradeVersion
			bookkeeper.Spec.Options["minorCompactionThreshold"] = "0.5"
			bookkeeper.Spec.JVMOptions.GcOpts = gcOpts

			// updating bookkeepercluster
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// checking if the upgrade of options was successful
			err = bookkeeper_e2eutil.WaitForCMBKClusterToUpgrade(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest bookkeeper cluster object

			// This is to get the latest Bookkeeper cluster object
			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			err = bookkeeper_e2eutil.CheckConfigMap(&t, k8sClient, bookkeeper, "BOOKIE_GC_OPTS", gcOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Spec.Version).To(Equal(upgradeVersion))
			Expect(bookkeeper.Spec.Options["minorCompactionThreshold"]).To(Equal("0.5"))

			// updating non-modifiable bookkeeper option journalDirectories
			bookkeeper.Spec.Options["journalDirectories"] = "journal"

			//updating bookkeepercluster
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(strings.ContainsAny(err.Error(), "path of journal directories should not be changed")).To(Equal(true))

			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Spec.Options["journalDirectories"]).To(Equal("/bk/journal"))

			// updating non-modifiable bookkeeper option useHostNameAsBookieID
			bookkeeper.Spec.Options["useHostNameAsBookieID"] = "false"

			//updating bookkeepercluster
			err = bookkeeper_e2eutil.UpdateBKCluster(&t, k8sClient, bookkeeper)
			Expect(strings.ContainsAny(err.Error(), "value of useHostNameAsBookieID should not be changed")).To(Equal(true))

			bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Spec.Options["useHostNameAsBookieID"]).To(Equal("true"))

			// Delete cluster
			err = bookkeeper_e2eutil.DeleteBKCluster(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(&t, k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
