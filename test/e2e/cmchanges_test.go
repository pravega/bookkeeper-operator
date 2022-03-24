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

	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Test create and recreate Bookkeeper cluster with the same name", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)

	BeforeEach(func() {
		defaultCluster = e2eutil.NewDefaultCluster(namespace)
	})

	Context("Creating a bookkeeper cluster", func() {
		var (
			bookkeeper *v1alpha1.BookkeeperCluster
			err        error
		)
		initialVersion := "0.6.0"
		upgradeVersion := "0.7.0"
		gcOpts := []string{"-XX:+UseG1GC", "-XX:MaxGCPauseMillis=10"}
		gcOptions := strings.Join(gcOpts, " ")

		BeforeEach(func() {
			defaultCluster.WithDefaults()
			defaultCluster.Spec.Version = initialVersion
			defaultCluster.Spec.Options["minorCompactionThreshold"] = "0.4"
			defaultCluster.Spec.Options["journalDirectories"] = "/bk/journal"
			defaultCluster.Spec.Options["useHostNameAsBookieID"] = "true"
			defaultCluster.Spec.JVMOptions.GcOpts = gcOpts
		})

		It("Should create successfully", func() {
			bookkeeper, err = e2eutil.CreateBKCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForBookkeeperClusterToBecomeReady(k8sClient, bookkeeper), timeout).Should(Succeed())
		})
		It("should have the proper configmap", func() {
			// This is to get the latest bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			err = e2eutil.CheckConfigMap(k8sClient, bookkeeper, "BOOKIE_GC_OPTS", gcOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should update the spec successfully", func() {
			// updating modifiable bookkeeper option
			gcOpts = []string{"-XX:-UseParallelGC", "-XX:MaxGCPauseMillis=10"}
			gcOptions = strings.Join(gcOpts, " ")
			bookkeeper.Spec.Version = upgradeVersion
			bookkeeper.Spec.Options["minorCompactionThreshold"] = "0.5"
			bookkeeper.Spec.JVMOptions.GcOpts = gcOpts

			// updating bookkeepercluster
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())

			// checking if the upgrade of options was successful
			Eventually(e2eutil.WaitForCMBKClusterToUpgrade(k8sClient, bookkeeper)).Should(Succeed())
		})

		It("should not update certain spec items", func() {
			// This is to get the latest bookkeeper cluster object
			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			err = e2eutil.CheckConfigMap(k8sClient, bookkeeper, "BOOKIE_GC_OPTS", gcOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Spec.Version).To(Equal(upgradeVersion))
			Expect(bookkeeper.Spec.Options["minorCompactionThreshold"]).To(Equal("0.5"))

			// updating non-modifiable bookkeeper option journalDirectories
			bookkeeper.Spec.Options["journalDirectories"] = "journal"

			// updating bookkeepercluster
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(strings.ContainsAny(err.Error(), "path of journal directories should not be changed")).To(Equal(true))

			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Spec.Options["journalDirectories"]).To(Equal("/bk/journal"))

			// updating non-modifiable bookkeeper option useHostNameAsBookieID
			bookkeeper.Spec.Options["useHostNameAsBookieID"] = "false"

			// updating bookkeepercluster
			err = e2eutil.UpdateBKCluster(k8sClient, bookkeeper)
			Expect(strings.ContainsAny(err.Error(), "value of useHostNameAsBookieID should not be changed")).To(Equal(true))

			bookkeeper, err = e2eutil.GetBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Expect(bookkeeper.Spec.Options["useHostNameAsBookieID"]).To(Equal("true"))
		})
		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeleteBKCluster(k8sClient, bookkeeper)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForBKClusterToTerminate(k8sClient, bookkeeper), timeout).Should(Succeed())
		})

	})
})
