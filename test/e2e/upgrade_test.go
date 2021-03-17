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
	"testing"

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

func testUpgradeCluster(t *testing.T) {
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
	firstUpgradeVersion := "0.7.0"
	secondUpgradeVersion := "0.7.1"
	cluster.Spec.Version = initialVersion
	cluster.Spec.Image.PullPolicy = "IfNotPresent"

	bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(t, f, ctx, cluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default Bookkeeper cluster should have 3 pods
	podSize := 3
	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bookkeeper, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Bookkeeper cluster object
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bookkeeper.Status.CurrentVersion).To(Equal(initialVersion))

	// This is to get the latest Bookkeeper cluster object
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	bookkeeper.Spec.Version = firstUpgradeVersion
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// trigger another upgrade while this upgrade is happening- it should fail
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())
	bookkeeper.Spec.Version = secondUpgradeVersion
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	fmt.Printf("\nBookkeeper cluster \n %+v", bookkeeper)
	fmt.Printf("\n --- \nError \n%+v", err)
	g.Expect(err).To(HaveOccurred(), "Should reject upgrade request while upgrade is in progress")
	g.Expect(err.Error()).To(ContainSubstring("failed to process the request, cluster is upgrading"))

	err = bookkeeper_e2eutil.WaitForBKClusterToUpgrade(t, f, ctx, bookkeeper, firstUpgradeVersion)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Bookkeeper cluster object
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(bookkeeper.Spec.Version).To(Equal(firstUpgradeVersion))
	g.Expect(bookkeeper.Status.CurrentVersion).To(Equal(firstUpgradeVersion))
	g.Expect(bookkeeper.Status.TargetVersion).To(Equal(""))

	// check version history
	g.Expect(bookkeeper.Status.VersionHistory[0]).To(Equal("0.6.0"))
	g.Expect(bookkeeper.Status.VersionHistory[1]).To(Equal("0.7.0"))

	// Delete cluster
	err = bookkeeper_e2eutil.DeleteBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

}
