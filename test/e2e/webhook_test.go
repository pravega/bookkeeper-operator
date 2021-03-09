/**
 * Copyright (c) 2019 Dell Inc., or its subsidiaries. All Rights Reserved.
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

func testWebhook(t *testing.T) {
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

	//Test webhook with an unsupported Bookkeeper cluster version
	invalidVersion := bookkeeper_e2eutil.NewClusterWithVersion(namespace, "99.0.0")
	invalidVersion.WithDefaults()
	_, err = bookkeeper_e2eutil.CreateBKCluster(t, f, ctx, invalidVersion)
	g.Expect(err).To(HaveOccurred(), "Should reject deployment of unsupported version")
	g.Expect(err.Error()).To(ContainSubstring("unsupported Bookkeeper cluster version 99.0.0"))

	// Test webhook with a supported Bookkeeper cluster version
	validVersion := bookkeeper_e2eutil.NewClusterWithVersion(namespace, "0.6.0")
	validVersion.WithDefaults()
	bookkeeper, err := bookkeeper_e2eutil.CreateBKCluster(t, f, ctx, validVersion)
	g.Expect(err).NotTo(HaveOccurred())

	podSize := 3
	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bookkeeper, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// Try to upgrade to a non-supported version
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	bookkeeper.Spec.Version = "99.0.0"
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).To(HaveOccurred(), "Should reject upgrade to an unsupported version")
	g.Expect(err.Error()).To(ContainSubstring("unsupported Bookkeeper cluster version 99.0.0"))

	// Try to downgrade the cluster
	bookkeeper.Spec.Version = "0.5.0"
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).To(HaveOccurred(), "Should not allow downgrade")
	g.Expect(err.Error()).To(ContainSubstring("unsupported upgrade from version 0.6.0 to 0.5.0"))

	// Delete cluster
	err = bookkeeper_e2eutil.DeleteBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())
}
