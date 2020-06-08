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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBookkeepercluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravegacluster")
}

var _ = Describe("bookkeepercluster", func() {

	Context("PdbNameForBookie", func() {
		var str1 string
		BeforeEach(func() {
			str1 = PdbNameForBookie("bk")
		})
		It("should return pdb name", func() {
			Ω(str1).To(Equal("bk-bookie"))
		})

	})
	Context("ConfigMapNameForBookie", func() {
		var str1 string
		BeforeEach(func() {
			str1 = ConfigMapNameForBookie("bk")
		})
		It("should return configmap name", func() {
			Ω(str1).To(Equal("bk-bookie"))
		})

	})
	Context("StatefulSetNameForBookie", func() {
		var str1 string
		BeforeEach(func() {
			str1 = StatefulSetNameForBookie("bk")
		})
		It("should return statefulset name", func() {
			Ω(str1).To(Equal("bk-bookie"))
		})
	})
	Context("HeadlessServiceNameForBookie", func() {
		var str1 string
		BeforeEach(func() {
			str1 = HeadlessServiceNameForBookie("bk")
		})
		It("should return headless service name", func() {
			Ω(str1).To(Equal("bk-bookie-headless"))
		})
	})
	Context("LabelsForBookie", func() {
		var str1 map[string]string
		var bk *v1alpha1.BookkeeperCluster
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			str1 = LabelsForBookie(bk)
		})
		It("should return label for app", func() {
			Ω(str1["app"]).To(Equal("bookkeeper-cluster"))
		})
		It("should return label for cluster name", func() {
			Ω(str1["bookkeeper_cluster"]).To(Equal("default"))
		})
		It("should return label for component", func() {
			Ω(str1["component"]).To(Equal("bookie"))
		})
	})
	Context("IsOrphan", func() {
		var result1, result2, result3, result4 bool
		BeforeEach(func() {

			result1 = IsOrphan("bookie-4", 3)
			result2 = IsOrphan("bookie-2", 3)
			result3 = IsOrphan("bookie", 1)
			result4 = IsOrphan("bookie-1ab", 1)
		})
		It("should return true for result2", func() {
			Ω(result1).To(Equal(true))
		})
		It("should return false for result1", func() {
			Ω(result2).To(Equal(false))
		})
		It("should return false for result3", func() {
			Ω(result3).To(Equal(false))
		})
		It("should return false for result4", func() {
			Ω(result4).To(Equal(false))
		})
	})
	Context("HealthcheckCommand()", func() {

		out := HealthcheckCommand(1234)
		It("should not be nil", func() {
			Ω(len(out)).ShouldNot(Equal(0))
		})

	})
	Context("Min()", func() {

		It("Min should be 10", func() {
			Ω(Min(10, 20)).Should(Equal(int32(10)))
		})
		It("Min should be 20", func() {
			Ω(Min(30, 20)).Should(Equal(int32(20)))
		})

	})
	Context("ContainsStringWithPrefix", func() {
		var result, result1 bool
		BeforeEach(func() {
			opts := []string{
				"-Xms512m",
				"-XX:+ExitOnOutOfMemoryError",
			}

			result = ContainsStringWithPrefix(opts, "-Xms")
			result1 = ContainsStringWithPrefix(opts, "-abc")
		})
		It("should return true for result", func() {
			Ω(result).To(Equal(true))

		})
		It("should return false for result1", func() {
			Ω(result1).To(Equal(false))
		})

	})
	Context("RemoveString", func() {
		var result bool
		BeforeEach(func() {
			opts := []string{
				"abc-test",
				"test1",
			}
			opts = RemoveString(opts, "abc-test")
			result = ContainsStringWithPrefix(opts, "abc")

		})
		It("should return false for result", func() {
			Ω(result).To(Equal(false))

		})

	})
	Context("GetStringWithPrefix", func() {
		var out, out1 string
		BeforeEach(func() {
			opts := []string{
				"abc-test",
				"test1",
			}
			out = GetStringWithPrefix(opts, "abc")
			out1 = GetStringWithPrefix(opts, "bk")

		})
		It("should return string with prefix", func() {
			Ω(out).To(Equal("abc-test"))

		})
		It("should return empty string", func() {
			Ω(out1).To(Equal(""))
		})
	})
	Context("GetClusterExpectedSize", func() {
		var replicas int
		var bk *v1alpha1.BookkeeperCluster
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			bk.WithDefaults()
			replicas = GetClusterExpectedSize(bk)
		})

		It("should return correct replica count", func() {
			Ω(replicas).To(Equal(3))
		})
	})
	Context("BookkeeperImage", func() {
		var image string
		var bk *v1alpha1.BookkeeperCluster
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			bk.WithDefaults()
			image = BookkeeperImage(bk)
		})

		It("should return correct image", func() {
			Ω(image).To(Equal("pravega/bookkeeper:0.6.1"))

		})
	})
	Context("BookkeeperTargetImage", func() {
		var image, image1 string
		var bk *v1alpha1.BookkeeperCluster
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			bk.WithDefaults()
			image, _ = BookkeeperTargetImage(bk)
			bk.Status.Init()
			bk.Status.TargetVersion = "0.6.1"
			image1, _ = BookkeeperTargetImage(bk)
		})
		It("should return empty", func() {
			Ω(image).To(Equal(""))
		})
		It("should return correct image", func() {
			Ω(image1).To(Equal("pravega/bookkeeper:0.6.1"))
		})
	})
	Context("ContainsVersion fn", func() {
		var result1, result2, result3 bool
		BeforeEach(func() {
			input := []string{"0.4.0", "0.5.0", "a.b.c"}
			result1 = ContainsVersion(input, "0.4.0")
			result2 = ContainsVersion(input, "0.7.0")
			result3 = ContainsVersion(input, "")

		})
		It("should return true for result", func() {
			Ω(result1).To(Equal(true))
		})
		It("should return false for result", func() {
			Ω(result2).To(Equal(false))
		})
		It("should return false for result", func() {
			Ω(result3).To(Equal(false))
		})
	})

	Context("GetPodVersion", func() {
		var out string
		BeforeEach(func() {
			annotationsMap := map[string]string{
				"bookkeeper.version": "0.7.0",
			}
			testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Annotations: annotationsMap}}
			out = GetPodVersion(testpod)
		})
		It("should return correct version", func() {
			Ω(out).To(Equal("0.7.0"))
		})
	})
	Context("OverrideDefaultJVMOptions", func() {
		var result, result1 []string
		BeforeEach(func() {
			jvmOpts := []string{
				"-Xms512m",
				"-XX:+ExitOnOutOfMemoryError",
				"-XX:+CrashOnOutOfMemoryError",
				"-XX:+HeapDumpOnOutOfMemoryError",
				"-XX:HeapDumpPath=/heap",
			}
			customOpts := []string{
				"-Xms1024m",
				"-XX:+ExitOnOutOfMemoryError",
				"-XX:+CrashOnOutOfMemoryError",
				"-XX:+HeapDumpOnOutOfMemoryError",
				"-XX:HeapDumpPath=/heap",
				"-yy:mem",
				"",
			}

			result = OverrideDefaultJVMOptions(jvmOpts, customOpts)
			result1 = OverrideDefaultJVMOptions(jvmOpts, result1)

		})
		It("should contain updated string", func() {
			Ω(len(result)).ShouldNot(Equal(0))
			Ω(result[0]).To(Equal("-Xms1024m"))
			Ω(result1[0]).To(Equal("-Xms512m"))
		})
	})
})
