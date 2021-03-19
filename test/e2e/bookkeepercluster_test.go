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

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	apis "github.com/pravega/bookkeeper-operator/pkg/apis"
	operator "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	bookkeeper_e2eutil "github.com/pravega/bookkeeper-operator/pkg/test/e2e/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBookkeeperCluster(t *testing.T) {
	bookkeeperClusterList := &operator.BookkeeperClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BookkeeperCluster",
			APIVersion: "pravega.pravega.io/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, bookkeeperClusterList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("x", testBookkeeperCluster)
}

func testBookkeeperCluster(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: bookkeeper_e2eutil.CleanupTimeout, RetryInterval: bookkeeper_e2eutil.CleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for bookkeeper-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "bookkeeper-operator", 1, bookkeeper_e2eutil.RetryInterval, bookkeeper_e2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}

	testFuncs := map[string]func(t *testing.T){
		"testDeletePods":            testDeletePods,
		"testCreateRecreateCluster": testCreateRecreateCluster,
		"testMultiBKCluster":        testMultiBKCluster,
		"testScaleCluster":          testScaleCluster,
		"testUpgradeCluster":        testUpgradeCluster,
		"testRollbackCluster":       testRollbackCluster,
		"testWebhook":               testWebhook,
		"testCMUpgradeCluster":      testCMUpgradeCluster,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}
