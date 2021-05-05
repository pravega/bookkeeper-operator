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
	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
	"strconv"
	"testing"
	"time"
)

func testMultiBKCluster(t *testing.T) {
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

	// Create first cluster
	cluster := bookkeeper_e2eutil.NewDefaultCluster(namespace)
	cm_name := "configmap1"
	cm1 := bookkeeper_e2eutil.NewConfigMap(namespace, cm_name, "pr1")
	err = bookkeeper_e2eutil.CreateConfigMap(t, f, ctx, cm1)
	g.Expect(err).NotTo(HaveOccurred())
	cluster.ObjectMeta.Name = "bk1"
	autorecovery := true
	cluster.Spec.AutoRecovery = &(autorecovery)
	cluster.WithDefaults()

	bk1, err := bookkeeper_e2eutil.CreateBKClusterWithCM(t, f, ctx, cluster, cm_name)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bk1)
	g.Expect(err).NotTo(HaveOccurred())

	bk1, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bk1)
	g.Expect(err).NotTo(HaveOccurred())
	err = bookkeeper_e2eutil.CheckConfigMap(t, f, ctx, bk1, "BK_autoRecoveryDaemonEnabled", strconv.FormatBool(autorecovery))
	g.Expect(err).NotTo(HaveOccurred())

	// Create second cluster
	cluster = bookkeeper_e2eutil.NewDefaultCluster(namespace)
	cm_name = "configmap2"
	cm2 := bookkeeper_e2eutil.NewConfigMap(namespace, cm_name, "pr2")
	err = bookkeeper_e2eutil.CreateConfigMap(t, f, ctx, cm2)
	g.Expect(err).NotTo(HaveOccurred())
	cluster.ObjectMeta.Name = "bk2"
	autorecovery = false
	cluster.Spec.AutoRecovery = &(autorecovery)
	cluster.WithDefaults()

	bk2, err := bookkeeper_e2eutil.CreateBKClusterWithCM(t, f, ctx, cluster, cm_name)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bk2)
	g.Expect(err).NotTo(HaveOccurred())

	bk2, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bk2)
	g.Expect(err).NotTo(HaveOccurred())
	err = bookkeeper_e2eutil.CheckConfigMap(t, f, ctx, bk2, "BK_autoRecoveryDaemonEnabled", strconv.FormatBool(autorecovery))
	g.Expect(err).NotTo(HaveOccurred())

	// Create third cluster
	cluster = bookkeeper_e2eutil.NewDefaultCluster(namespace)
	cluster.WithDefaults()

	bk3, err := bookkeeper_e2eutil.CreateBKCluster(t, f, ctx, cluster)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bk3)
	g.Expect(err).NotTo(HaveOccurred())

	bk3, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bk3)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Bookkeeper cluster object
	bk1, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bk1)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale up replicas in the first Bookkeeper cluster
	bk1.Spec.Replicas = 5

	err = bookkeeper_e2eutil.UpdateBKCluster(t, f, ctx, bk1)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bk1)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Bookkeeper cluster object
	bk2, err = bookkeeper_e2eutil.GetBKCluster(t, f, ctx, bk2)
	g.Expect(err).NotTo(HaveOccurred())

	// Deleting pods of the second Bookkeeper cluster
	podDeleteCount := 3
	err = bookkeeper_e2eutil.DeletePods(t, f, ctx, bk2, podDeleteCount)
	g.Expect(err).NotTo(HaveOccurred())
	time.Sleep(10 * time.Second)

	err = bookkeeper_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, bk2)
	g.Expect(err).NotTo(HaveOccurred())

	// deleting all bookkeeper clusters
	err = bookkeeper_e2eutil.DeleteBKCluster(t, f, ctx, bk1)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, bk1)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.DeleteBKCluster(t, f, ctx, bk2)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, bk2)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.DeleteBKCluster(t, f, ctx, bk3)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, bk3)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.DeleteConfigMap(t, f, ctx, cm1)
	g.Expect(err).NotTo(HaveOccurred())

	err = bookkeeper_e2eutil.DeleteConfigMap(t, f, ctx, cm2)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

}
