/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2eutil

import (
	api "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDefaultCluster returns a cluster with an empty spec, which will be filled
// with default values
func NewDefaultCluster(namespace string) *api.BookkeeperCluster {
	return &api.BookkeeperCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BookkeeperCluster",
			APIVersion: "pravega.pravega.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bookkeeper",
			Namespace: namespace,
		},
		Spec: api.BookkeeperClusterSpec{},
	}
}

func NewClusterWithVersion(namespace, version string) *api.BookkeeperCluster {
	cluster := NewDefaultCluster(namespace)
	cluster.Spec = api.BookkeeperClusterSpec{
		Version: version,
	}
	return cluster
}

func NewConfigMap(namespace, name string, pravega string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			"PRAVEGA_CLUSTER_NAME": pravega,
		},
	}
}
