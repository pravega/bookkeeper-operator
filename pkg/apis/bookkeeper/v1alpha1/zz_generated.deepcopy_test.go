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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
)

var _ = Describe("DeepCopy", func() {
	Context("with defaults", func() {
		var str1, str2 string
		var str3, str4 v1.PullPolicy
		var bk1, bk2 *v1alpha1.BookkeeperCluster

		BeforeEach(func() {
			bk1 = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			bk1.WithDefaults()
			bk1.Status.Init()
			bk1.Status.VersionHistory = []string{"0.6.0", "0.7.0"}
			bk1.Spec.Options["key"] = "value"
			temp := *bk1.DeepCopy()
			bk2 = &temp
			str1 = bk1.Spec.Image.Repository
			str2 = bk2.Spec.Image.Repository
			str3 = bk1.Spec.Image.PullPolicy
			str4 = bk2.Spec.Image.PullPolicy
			bk1.Spec.Image.PullPolicy = "PullIfNotPresent"
			bk1.Spec.Image.DeepCopyInto(bk2.Spec.Image)
			bk1.Spec.Image.Repository = "bk/exmple"
			bk1.Spec.Probes.ReadinessProbe.InitialDelaySeconds = 5
			bk1.Spec.Probes.LivenessProbe.FailureThreshold = 2
			bk2.Spec.Probes = bk1.Spec.Probes.DeepCopy()
			bk1.Spec.JVMOptions.MemoryOpts = []string{"1g"}
			bk2.Spec.JVMOptions = bk1.Spec.JVMOptions.DeepCopy()
			bk2.Spec.Storage = bk1.Spec.Storage.DeepCopy()
			bk1.Spec.Options["ledgers"] = "l1"
			bk2.Spec = *bk1.Spec.DeepCopy()
			bk2.Status = *bk1.Status.DeepCopy()
			bk1.Status.Members.Ready = []string{"bookie-0", "bookie-1"}
			bk1.Status.Members.Unready = []string{"bookie-3", "bookie-2"}
			bk2.Status.Members = *bk1.Status.Members.DeepCopy()
			bk1.Spec.Image.Repository = "bookie/exmple1"
			bk2.Spec.Image = bk1.Spec.Image.DeepCopy()
			bk2.Spec.Image.ImageSpec = *bk1.Spec.Image.ImageSpec.DeepCopy()
			bk1.Status.SetPodsReadyConditionTrue()
			bk2.Status.Conditions[0] = *bk1.Status.Conditions[0].DeepCopy()
		})
		It("value of str1 and str2 should be equal", func() {
			Ω(str2).To(Equal(str1))
		})
		It("value of str3 and str4 should be equal", func() {
			Ω(str3).To(Equal(str4))
		})
		It("checking value of bk2 probes", func() {
			Ω(bk2.Spec.Probes.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(5)))
			Ω(bk2.Spec.Probes.LivenessProbe.FailureThreshold).To(Equal(int32(2)))
			bk1.Spec.Probes.ReadinessProbe.InitialDelaySeconds = 0
			bk1.Spec.Probes.ReadinessProbe.DeepCopyInto(bk2.Spec.Probes.ReadinessProbe)
			Ω(bk2.Spec.Probes.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(0)))
		})
		It("checking bk2 jvm option as 1g", func() {
			Ω(bk2.Spec.JVMOptions.MemoryOpts[0]).To(Equal("1g"))
		})
		It("checking bk2 options ledger field", func() {
			Ω(bk2.Spec.Options["ledgers"]).To(Equal("l1"))
		})
		It("checking bk2 ready members", func() {
			Ω(bk2.Status.Members.Ready[0]).To(Equal("bookie-0"))
		})
		It("checking bk2 unready members", func() {
			Ω(bk2.Status.Members.Unready[0]).To(Equal("bookie-3"))
		})
		It("checking bk2 spec image", func() {
			Ω(bk2.Spec.Image.Repository).To(Equal("bookie/exmple1"))
		})
		It("checking bk2 spec image", func() {
			Ω(bk2.Spec.Image.ImageSpec.Repository).To(Equal("bookie/exmple1"))
		})
		It("checking status conditions", func() {
			Ω(bk2.Status.Conditions[0].Reason).To(Equal(bk1.Status.Conditions[0].Reason))
		})
		It("checking bk2 spec storage", func() {
			Ω(bk2.Spec.Storage).To(Equal(bk1.Spec.Storage))
		})
		It("checking for nil member status", func() {
			var memberstatus *v1alpha1.MembersStatus
			memberstatus2 := memberstatus.DeepCopy()
			Ω(memberstatus2).To(BeNil())
		})
		It("checking for nil cluster status", func() {
			var clusterstatus *v1alpha1.BookkeeperClusterStatus
			clusterstatus2 := clusterstatus.DeepCopy()
			Ω(clusterstatus2).To(BeNil())
		})
		It("checking for nil cluster spec", func() {
			var clusterspec *v1alpha1.BookkeeperClusterSpec
			clusterspec2 := clusterspec.DeepCopy()
			Ω(clusterspec2).To(BeNil())
		})
		It("checking for nil cluster condition", func() {
			var clustercond *v1alpha1.ClusterCondition
			clustercond2 := clustercond.DeepCopy()
			Ω(clustercond2).To(BeNil())
		})
		It("checking for nil bookkeeper cluster", func() {
			var cluster *v1alpha1.BookkeeperCluster
			cluster2 := cluster.DeepCopy()
			Ω(cluster2).To(BeNil())
		})
		It("checking for nil imagespec", func() {
			var imagespec *v1alpha1.ImageSpec
			imagespec2 := imagespec.DeepCopy()
			Ω(imagespec2).To(BeNil())
		})
		It("checking for nil clusterlist", func() {
			var clusterlist *v1alpha1.BookkeeperClusterList
			clusterlist2 := clusterlist.DeepCopy()
			Ω(clusterlist2).To(BeNil())
		})
		It("checking for nil bookkeeper cluster deepcopyobject", func() {
			var cluster *v1alpha1.BookkeeperCluster
			cluster2 := cluster.DeepCopyObject()
			Ω(cluster2).To(BeNil())
		})
		It("checking for nil bookkeeper clusterlist deepcopyobject", func() {
			var clusterlist *v1alpha1.BookkeeperClusterList
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).To(BeNil())
		})
		It("checking for nil jvm options", func() {
			bk1.Spec.JVMOptions = nil
			Ω(bk1.Spec.JVMOptions.DeepCopy()).Should(BeNil())
		})
		It("checking for nil storage options", func() {
			bk1.Spec.Storage = nil
			Ω(bk1.Spec.Storage.DeepCopy()).Should(BeNil())
		})
		It("checking for nil bookkeeper image spec", func() {
			bk1.Spec.Image = nil
			Ω(bk1.Spec.Image.DeepCopy()).Should(BeNil())
		})
		It("checking for deepcopyobject for clusterlist", func() {
			var clusterlist v1alpha1.BookkeeperClusterList
			clusterlist.ResourceVersion = "v1alpha1"
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).ShouldNot(BeNil())
		})
		It("checking for deepcopyobject for clusterlist with items", func() {
			var clusterlist v1alpha1.BookkeeperClusterList
			clusterlist.ResourceVersion = "v1alpha1"
			clusterlist.Items = []v1alpha1.BookkeeperCluster{
				{
					Spec: v1alpha1.BookkeeperClusterSpec{},
				},
			}
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).ShouldNot(BeNil())
		})
		It("checking for deepcopy for clusterlist", func() {
			var clusterlist v1alpha1.BookkeeperClusterList
			clusterlist.ResourceVersion = "v1alpha1"
			clusterlist2 := clusterlist.DeepCopy()
			Ω(clusterlist2.ResourceVersion).To(Equal("v1alpha1"))
		})
		It("checking for Deepcopy object", func() {
			bk := bk2.DeepCopyObject()
			Ω(bk.GetObjectKind().GroupVersionKind().Version).To(Equal(""))
		})
	})
})
