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
	"time"

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	bkapi "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
	corev1 "k8s.io/api/core/v1"
)

func testRollbackCluster(t *testing.T) {
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
	firstUpgradeVersion := "0.7.0-1" // incorrect version
	secondUpgradeVersion := "0.5.0"
	cluster.Spec.Version = initialVersion
	cluster.Spec.Image.Repository = "pravega/bookkeeper"
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

	// waiting for upgrade to fail
	time.Sleep(2 * time.Minute)

	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())
	_, errorCondition := bookkeeper.Status.GetClusterCondition(bkapi.ClusterConditionError)
	g.Expect(errorCondition.Status).To(Equal(corev1.ConditionTrue))
	g.Expect(errorCondition.Reason).To(Equal("UpgradeFailed"))
	g.Expect(errorCondition.Message).To(ContainSubstring("pod bookkeeper-bookie-0 update failed because of ImagePullBackOff"))

	// trigger rollback to version other than last stable version
	// expect failure
	bookkeeper.Spec.Version = secondUpgradeVersion
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).To(HaveOccurred(), "Should not allow rollback to any version other than the last stable version")
	g.Expect(err.Error()).To(ContainSubstring("Rollback to version 0.5.0 not supported. Only rollback to version 0.6.0 is supported"))

	// trigger rollback to last stable version
	bookkeeper.Spec.Version = initialVersion
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// trigger another upgrade while the last rollback is still ongoing
	// should be rejected by webhook
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())
	bookkeeper.Spec.Version = secondUpgradeVersion
	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).To(HaveOccurred(), "Should reject rollback request while rollback is in progress")
	g.Expect(err.Error()).To(ContainSubstring("failed to process the request, rollback in progress"))

	_, rollbackCondition := bookkeeper.Status.GetClusterCondition(bkapi.ClusterConditionRollback)
	g.Expect(rollbackCondition.Status).To(Equal(corev1.ConditionTrue))
	g.Expect(rollbackCondition.Reason).To(ContainSubstring("Updating Bookkeeper"))

	err = bookkeeper_e2eutil.WaitForBKClusterToRollback(t, f, ctx, bookkeeper, initialVersion)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Bookkeeper cluster object
	bookkeeper, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// wait for rollback to complete
	g.Expect(bookkeeper.Spec.Version).To(Equal(initialVersion))
	g.Expect(bookkeeper.Status.CurrentVersion).To(Equal(initialVersion))
	g.Expect(bookkeeper.Status.TargetVersion).To(Equal(""))

	// checking version history
	g.Expect(bookkeeper.Status.VersionHistory[0]).To(Equal("0.6.0"))

	// Delete cluster
	err = bookkeeper_e2eutil.DeleteBKCluster(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, bookkeeper)
	g.Expect(err).NotTo(HaveOccurred())

}
