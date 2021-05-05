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
	"testing"

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

func testCMUpgradeCluster(t *testing.T) {
	g := NewGomegaWithT(t)

	doCleanup := true
	ctx := framework.NewTestCtx(t)
	defer func() {
		if doCleanup {
			ctx.Cleanup()
		}
	}()

	namespace, err := ctx.GetNamespace()
	g.Expect(err).NotTo(HaveOccurred())
	f := framework.Global

	cluster := bookkeeper_e2eutil.NewDefaultCluster(namespace)

	cluster.WithDefaults()
	initialVersion := "0.6.0"
	upgradeVersion := "0.7.0"
	gcOpts := []string{"-XX:+UseG1GC", "-XX:MaxGCPauseMillis=10"}
	gcOptions := strings.Join(gcOpts, " ")
	cluster.Spec.Version = initialVersion
	cluster.Spec.Options["minorCompactionThreshold"] = "0.4"
	cluster.Spec.Options["journalDirectories"] = "/bk/journal"
	cluster.Spec.JVMOptions.GcOpts = gcOpts

	bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(t, f, ctx, cluster)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest bookkeeper cluster object
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())
	err = bookkeeper_e2eutil.CheckConfigMap(t, f, ctx, bookkeeper, "BOOKIE_GC_OPTS", gcOptions)
	g.Expect(err).NotTo(HaveOccurred())

	// updating bookkeeper option
	gcOpts = []string{"-XX:-UseParallelGC", "-XX:MaxGCPauseMillis=10"}
	gcOptions = strings.Join(gcOpts, " ")
	bookkeeper.Spec.Version = upgradeVersion
	bookkeeper.Spec.Options["minorCompactionThreshold"] = "0.5"
	bookkeeper.Spec.JVMOptions.GcOpts = gcOpts

	// updating bookkeepercluster
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// checking if the upgrade of options was successful
	err = bookkeeper_e2eutil.WaitForCMBKClusterToUpgrade(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest bookkeeper cluster object
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())
	err = bookkeeper_e2eutil.CheckConfigMap(t, f, ctx, bookkeeper, "BOOKIE_GC_OPTS", gcOptions)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bookkeeper.Spec.Version).To(Equal(upgradeVersion))
	g.Expect(bookkeeper.Spec.Options["minorCompactionThreshold"]).To(Equal("0.5"))

	// updating bookkeeper option
	bookkeeper.Spec.Options["journalDirectories"] = "journal"

	//updating bookkeepercluster
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(strings.ContainsAny(err.Error(), "path of journal directories should not be changed")).To(Equal(true))

	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bookkeeper.Spec.Options["journalDirectories"]).To(Equal("/bk/journal"))

	// Delete cluster
	err = bookkeeper_e2eutil.DeleteBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

}
