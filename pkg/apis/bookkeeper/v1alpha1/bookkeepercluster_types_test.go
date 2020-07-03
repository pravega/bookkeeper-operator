/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1alpha1_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
)

func TestV1alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BookkeeperCluster API")
}

var _ = Describe("BookkeeperCluster Types Spec", func() {

	var bk v1alpha1.BookkeeperCluster

	BeforeEach(func() {
		bk = v1alpha1.BookkeeperCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
	})

	Context("#WithDefaults", func() {
		var changed bool

		BeforeEach(func() {
			changed = bk.WithDefaults()
		})

		It("should return as changed", func() {
			Ω(changed).Should(BeTrue())
		})

		It("should set zookeeper uri", func() {
			Ω(bk.Spec.ZookeeperUri).Should(Equal("zookeeper-client:2181"))
		})

		It("should set version to 0.7.0", func() {
			Ω(bk.Spec.Version).Should(Equal("0.7.0"))
		})

	})
})
