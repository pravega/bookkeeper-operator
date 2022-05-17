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
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/bookkeeper-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

	Context("WithDefaults", func() {
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

		It("should set version to 0.11.0", func() {
			Ω(bk.Spec.Version).Should(Equal("0.11.0"))
		})
	})

	Context("NewEvent", func() {
		var bk *v1alpha1.BookkeeperCluster
		var event *corev1.Event
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			bk.WithDefaults()
			message := "upgrade failed"
			event = bk.NewEvent("bookie", "UPGRADE_ERROR", message, "Error")
		})

		It("Event size should not be zero", func() {
			Ω(event.Size()).ShouldNot(Equal(0))
		})
	})

	Context("NewApplicationEvent", func() {
		var bk *v1alpha1.BookkeeperCluster
		var event *corev1.Event
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			bk.WithDefaults()
			message := "upgrade failed"
			event = bk.NewApplicationEvent("bookie", "UPGRADE_ERROR", message, "Error")
		})
		It("Event size should not be zero", func() {
			Ω(event.Size()).ShouldNot(Equal(0))
		})
	})

	Context("WaitForClusterToTerminate", func() {
		var bk *v1alpha1.BookkeeperCluster
		var client client.Client
		var err error
		BeforeEach(func() {

			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			bk.WithDefaults()
			bk.Spec.Annotations = map[string]string{"bookkeeperAnnotation": "dummyAnnotation"}
			bk.Annotations = bk.AnnotationsForBookie()
			s := scheme.Scheme
			s.AddKnownTypes(v1alpha1.GroupVersion, bk)

			client = fake.NewFakeClient(bk)
			err = bk.WaitForClusterToTerminate(client)
		})
		It("should  be nil", func() {
			Ω(err).Should(BeNil())
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
			replicas = bk.GetClusterExpectedSize()
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
				Spec: v1alpha1.BookkeeperClusterSpec{
					Version: "0.6.1",
				},
			}
			bk.WithDefaults()
			image = bk.BookkeeperImage()
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
			image, _ = bk.BookkeeperTargetImage()
			bk.Status.Init()
			bk.Status.TargetVersion = "0.6.1"
			image1, _ = bk.BookkeeperTargetImage()
		})
		It("should return empty", func() {
			Ω(image).To(Equal(""))
		})
		It("should return correct image", func() {
			Ω(image1).To(Equal("pravega/bookkeeper:0.6.1"))
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
			bk.Spec.Labels = map[string]string{"bookkeeperLabel": "dummyLabel"}
			str1 = bk.LabelsForBookie()
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

	Context("ValidateDelete", func() {
		var (
			bk  *v1alpha1.BookkeeperCluster
			err error
		)
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			bk.WithDefaults()
			err = bk.ValidateDelete()

		})
		It("should return nil", func() {
			Ω(err).To(BeNil())
		})
	})

	Context("ValidateBookkeeperVersion", func() {
		var (
			bk  *v1alpha1.BookkeeperCluster
			err error
		)

		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			bk.WithDefaults()
		})

		Context("Spec version empty", func() {
			BeforeEach(func() {
				bk.Spec.Version = ""
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return nil", func() {
				Ω(err).To(BeNil())
			})
		})

		Context("Absolute Path is incorrect", func() {
			BeforeEach(func() {
				bk.Spec = v1alpha1.BookkeeperClusterSpec{
					Options: map[string]string{
						"journalDirectories": "bk/journal/j0,bk/journal/j1,bk/journal/j2,bk/journal/j3",
						"ledgerDirectories":  "bk/ledgers/l0,bk/ledgers/l1,bk/ledgers/l2,bk/ledgers/l3",
						"indexDirectories":   "",
					},
				}
				bk.WithDefaults()
				err = bk.ValidateAbsolutePath([]string{"journalDirectories", "ledgerDirectories", "indexDirectories"})
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Error validating absolute paths of journal/ledger/index directories")).Should(Equal(true))
			})
		})

		Context("Version not in valid format", func() {
			BeforeEach(func() {
				bk.Spec.Version = "999"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "request version is not in valid format")).Should(Equal(true))
			})
		})

		Context("Spec version and current version same", func() {
			BeforeEach(func() {
				bk.Spec.Version = "0.7.0"
				bk.Status.CurrentVersion = "0.7.0"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return nil", func() {
				Ω(err).To(BeNil())
			})
		})

		Context("current version not in correct format", func() {
			BeforeEach(func() {
				bk.Spec.Version = "0.7.0"
				bk.Status.CurrentVersion = "999"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "found version is not in valid format")).Should(Equal(true))
			})
		})

		Context("unsupported upgrade to a version", func() {
			BeforeEach(func() {
				bk.Status.CurrentVersion = "0.7.2"
				bk.Spec.Version = "0.7.0"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "downgrading the cluster from version 0.7.2 to 0.7.0 is not supported")).Should(Equal(true))
			})
		})

		Context("supported upgrade to a version", func() {
			BeforeEach(func() {
				bk.Status.CurrentVersion = "0.7.0"
				bk.Spec.Version = "0.7.3"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return nil", func() {
				Ω(err).To(BeNil())
			})
		})

		Context("validation while cluster upgrade in progress", func() {
			BeforeEach(func() {
				bk.Status.SetUpgradingConditionTrue(" ", " ")
				bk.Spec.Version = "0.7.1"
				bk.Status.TargetVersion = "0.7.0"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "failed to process the request, cluster is upgrading")).Should(Equal(true))
			})
		})

		Context("validation while cluster rollback in progress", func() {
			BeforeEach(func() {
				bk.Status.CurrentVersion = "0.7.0"
				bk.Status.Init()
				bk.Status.AddToVersionHistory("0.6.0")
				bk.Status.SetRollbackConditionTrue(" ", " ")
				bk.Spec.Version = "0.7.0"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "failed to process the request, rollback in progress")).Should(Equal(true))
			})
		})

		Context("validation while cluster in error state", func() {
			BeforeEach(func() {
				bk.Status.SetErrorConditionTrue("some err", " ")
				bk.Spec.Version = "0.7.0"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "failed to process the request, cluster is in error state")).Should(Equal(true))
			})
		})

		Context("validation while cluster in upgradefailed state", func() {
			BeforeEach(func() {
				bk.Status.CurrentVersion = "0.7.0"
				bk.Status.Init()
				bk.Status.AddToVersionHistory("0.6.0")
				bk.Status.SetErrorConditionTrue("UpgradeFailed", " ")
				bk.Spec.Version = "0.7.0"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Rollback to version 0.7.0 not supported")).Should(Equal(true))
			})
		})

		Context("validation while cluster in upgradefailed state and supported rollback version", func() {
			BeforeEach(func() {
				bk.Status.CurrentVersion = "0.6.0"
				bk.Status.Init()
				bk.Status.AddToVersionHistory("0.6.0")
				bk.Status.SetErrorConditionTrue("UpgradeFailed", " ")
				bk.Spec.Version = "0.6.0"
				err = bk.ValidateBookkeeperVersion()
			})
			It("should return nil", func() {
				Ω(err).To(BeNil())
			})
		})
	})

	Context("HeadlessServiceNameForBookie", func() {
		var str1 string
		BeforeEach(func() {
			bk.WithDefaults()
			str1 = bk.HeadlessServiceNameForBookie()
		})
		It("should return headless service name", func() {
			Ω(str1).To(Equal("default-bookie-headless"))
		})
	})

})
