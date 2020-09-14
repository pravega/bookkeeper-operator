/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */
package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("k8sutil", func() {
	Context("DownwardAPIEnv()", func() {
		env := DownwardAPIEnv()
		It("should not be nil", func() {
			Ω(env).ShouldNot(BeNil())
		})
	})
	Context("PodAntiAffinity", func() {
		affinity := PodAntiAffinity("bookie", "bkcluster")
		It("should not be nil", func() {
			Ω(affinity).ShouldNot(BeNil())
		})

	})
	Context("podReady", func() {
		var result, result1 bool
		BeforeEach(func() {
			testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
				Status: v1.PodStatus{
					Conditions: []v1.PodCondition{
						{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						},
					}},
			}
			testpod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}}}
			result = IsPodReady(testpod)
			result1 = IsPodReady(testpod1)
		})
		It("pod ready should be true", func() {
			Ω(result).To(Equal(true))
		})
		It("pod ready should be false", func() {
			Ω(result1).To(Equal(false))
		})
	})
	Context("podFaulty", func() {
		var result, result1 bool
		BeforeEach(func() {
			testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "test",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
						},
					}},
			}
			testpod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "test",
							State: v1.ContainerState{},
						},
					}},
			}
			result, _ = IsPodFaulty(testpod)
			result1, _ = IsPodFaulty(testpod1)
		})
		It("pod faulty should be true", func() {
			Ω(result).To(Equal(true))
		})
		It("pod faulty should be false", func() {
			Ω(result1).To(Equal(false))
		})
	})

})
