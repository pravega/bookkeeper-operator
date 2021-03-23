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
	"testing"

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
)

func testScaleCluster(t *testing.T) {
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

	defaultCluster := bookkeeper_e2eutil.NewDefaultCluster(namespace)
	defaultCluster.WithDefaults()

	bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Bookkeeper cluster object
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale up Bookkeeper cluster
	bookkeeper.Spec.Replicas = 5

	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Bookkeeper cluster object
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale down Bookkeeper cluster back to default
	bookkeeper.Spec.Replicas = 3

	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// Delete cluster
	err = bookkeeper_e2eutil.DeleteBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

}
