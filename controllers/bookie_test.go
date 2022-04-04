/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package controllers_test

import (
	"fmt"
	"strings"

	"github.com/pravega/bookkeeper-operator/api/v1alpha1"

	bookkeepercluster "github.com/pravega/bookkeeper-operator/controllers"
	"github.com/pravega/bookkeeper-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bookie", func() {
	var _ = Describe("Bookie Test", func() {
		var (
			bk *v1alpha1.BookkeeperCluster
		)
		BeforeEach(func() {
			bk = &v1alpha1.BookkeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
		})

		Context("User is specifying bookkeeper journal and ledger path", func() {
			var (
				customReq *corev1.ResourceRequirements
			)
			BeforeEach(func() {
				customReq = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("6Gi"),
					},
				}
				boolFalse := false
				bk.Spec = v1alpha1.BookkeeperClusterSpec{
					Version:            "0.4.0",
					ServiceAccountName: "bk-operator",
					EnvVars:            "bk-configmap",
					AutoRecovery:       &boolFalse,
					Resources:          customReq,
					Probes: &v1alpha1.Probes{
						ReadinessProbe: &v1alpha1.Probe{
							InitialDelaySeconds: 10,
							PeriodSeconds:       5,
							FailureThreshold:    5,
							SuccessThreshold:    1,
							TimeoutSeconds:      2,
						},
						LivenessProbe: &v1alpha1.Probe{
							InitialDelaySeconds: 10,
							PeriodSeconds:       5,
							FailureThreshold:    5,
							SuccessThreshold:    1,
							TimeoutSeconds:      2,
						},
					},
					JVMOptions: &v1alpha1.JVMOptions{
						MemoryOpts:    []string{"-XX:MaxDirectMemorySize=2g"},
						GcOpts:        []string{"-XX:+UseG1GC", "-XX:MaxGCPauseMillis=10"},
						GcLoggingOpts: []string{},
						ExtraOpts:     []string{"-XX:+IgnoreUnrecognizedVMOptions"},
					},
					Options: map[string]string{
						"journalDirectories":    "/bk/journal/j0,/bk/journal/j1,/bk/journal/j2,/bk/journal/j3",
						"ledgerDirectories":     "/bk/ledgers/l0,/bk/ledgers/l1,/bk/ledgers/l2,/bk/ledgers/l3",
						"indexDirectories":      "/bk/index/i0,/bk/index/i1",
						"hostPathVolumeMounts":  "foo=/tmp/foo,bar=/tmp/bar",
						"emptyDirVolumeMounts":  "baz=/tmp/baz,quux=/tmp/quux",
						"configMapVolumeMounts": "bk-log4j:log4j.properties=/opt/bookkeeper/conf/log4j.properties",
					},
					Labels: map[string]string{
						"bookie-name": "dummyBookie",
					},
					Annotations: map[string]string{
						"bookie-annotation": "dummyBookie",
					},
					InitContainers: []corev1.Container{
						corev1.Container{
							Name:    "testing",
							Image:   "dummy-image",
							Command: []string{"sh", "-c", "ls;pwd"},
						},
					},
					RunAsPrivilegedUser: &boolFalse,
				}
				bk.WithDefaults()
			})
			Context("Bookkeeper", func() {
				It("should create a headless service", func() {
					headlessservice := bookkeepercluster.MakeBookieHeadlessService(bk)
					Ω(headlessservice.Name).Should(Equal(bk.HeadlessServiceNameForBookie()))
				})

				It("should create a pod disruption budget", func() {
					pdb := bookkeepercluster.MakeBookiePodDisruptionBudget(bk)
					Ω(pdb.Name).Should(Equal(util.PdbNameForBookie(bk.Name)))
				})

				It("should create a config-map", func() {
					cm := bookkeepercluster.MakeBookieConfigMap(bk)
					Ω(cm.Name).Should(Equal(util.ConfigMapNameForBookie(bk.Name)))
				})

				It("should create a stateful set", func() {
					ss := bookkeepercluster.MakeBookieStatefulSet(bk)
					Ω(ss.Name).Should(Equal(util.StatefulSetNameForBookie(bk.Name)))
				})

				It("should create a stateful set", func() {
					ss := bookkeepercluster.MakeBookieStatefulSet(bk)
					Ω(ss.Labels["bookie-name"]).Should(Equal("dummyBookie"))
					Ω(ss.Annotations["bookie-annotation"]).Should(Equal("dummyBookie"))
				})

				It("should set the JVM options given by user", func() {
					cm := bookkeepercluster.MakeBookieConfigMap(bk)
					memoryOpts := cm.Data["BOOKIE_MEM_OPTS"]
					gcOpts := cm.Data["BOOKIE_GC_OPTS"]
					gcLoggingOpts := cm.Data["BOOKIE_GC_LOGGING_OPTS"]
					extraOpts := cm.Data["BOOKIE_EXTRA_OPTS"]
					Ω(memoryOpts).Should(Equal("-XX:MaxDirectMemorySize=2g"))
					Ω(gcOpts).Should(Equal("-XX:+UseG1GC -XX:MaxGCPauseMillis=10"))
					Ω(gcLoggingOpts).Should(Equal(""))
					Ω(extraOpts).Should(Equal("-XX:+IgnoreUnrecognizedVMOptions"))
				})

				It("should have journal and ledgers dir set to the values given by user", func() {
					sts := bookkeepercluster.MakeBookieStatefulSet(bk)
					mountledger0 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath
					mountledger1 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath
					mountledger2 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[2].MountPath
					mountledger3 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[3].MountPath
					Ω(mountledger0).Should(Equal("/bk/ledgers/l0"))
					Ω(mountledger1).Should(Equal("/bk/ledgers/l1"))
					Ω(mountledger2).Should(Equal("/bk/ledgers/l2"))
					Ω(mountledger3).Should(Equal("/bk/ledgers/l3"))
					mountjournal0 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[4].MountPath
					mountjournal1 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[5].MountPath
					mountjournal2 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[6].MountPath
					mountjournal3 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[7].MountPath
					Ω(mountjournal0).Should(Equal("/bk/journal/j0"))
					Ω(mountjournal1).Should(Equal("/bk/journal/j1"))
					Ω(mountjournal2).Should(Equal("/bk/journal/j2"))
					Ω(mountjournal3).Should(Equal("/bk/journal/j3"))
					mountindex0 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[8].MountPath
					mountindex1 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[9].MountPath
					Ω(mountindex0).Should(Equal("/bk/index/i0"))
					Ω(mountindex1).Should(Equal("/bk/index/i1"))
				})

				It("should have hostPathVolumeMounts set to the values given by user", func() {
					sts := bookkeepercluster.MakeBookieStatefulSet(bk)
					mounthostpath0 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[10].MountPath
					mounthostpath1 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[11].MountPath
					Ω(mounthostpath0).Should(Equal("/tmp/foo"))
					Ω(mounthostpath1).Should(Equal("/tmp/bar"))
				})

				It("should have emptyDirVolumeMounts set to the values given by user", func() {
					sts := bookkeepercluster.MakeBookieStatefulSet(bk)
					mounthostpath0 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[12].MountPath
					mounthostpath1 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[13].MountPath
					Ω(mounthostpath0).Should(Equal("/tmp/baz"))
					Ω(mounthostpath1).Should(Equal("/tmp/quux"))
				})

				It("should have configMapVolumeMounts set to the values given by user", func() {
					sts := bookkeepercluster.MakeBookieStatefulSet(bk)
					mounthostpath0 := sts.Spec.Template.Spec.Containers[0].VolumeMounts[14].MountPath
					Ω(mounthostpath0).Should(Equal("/opt/bookkeeper/conf/log4j.properties"))
				})

				It("should have probe timeout values set to the values given by user", func() {
					rp_i := bk.Spec.Probes.ReadinessProbe.InitialDelaySeconds
					rp_p := bk.Spec.Probes.ReadinessProbe.PeriodSeconds
					rp_f := bk.Spec.Probes.ReadinessProbe.FailureThreshold
					rp_s := bk.Spec.Probes.ReadinessProbe.SuccessThreshold
					rp_t := bk.Spec.Probes.ReadinessProbe.TimeoutSeconds
					Ω(rp_i).Should(Equal(int32(10)))
					Ω(rp_p).Should(Equal(int32(5)))
					Ω(rp_f).Should(Equal(int32(5)))
					Ω(rp_s).Should(Equal(int32(1)))
					Ω(rp_t).Should(Equal(int32(2)))
					lp_i := bk.Spec.Probes.LivenessProbe.InitialDelaySeconds
					lp_p := bk.Spec.Probes.LivenessProbe.PeriodSeconds
					lp_f := bk.Spec.Probes.LivenessProbe.FailureThreshold
					lp_s := bk.Spec.Probes.LivenessProbe.SuccessThreshold
					lp_t := bk.Spec.Probes.LivenessProbe.TimeoutSeconds
					Ω(lp_i).Should(Equal(int32(10)))
					Ω(lp_p).Should(Equal(int32(5)))
					Ω(lp_f).Should(Equal(int32(5)))
					Ω(lp_s).Should(Equal(int32(1)))
					Ω(lp_t).Should(Equal(int32(2)))
				})

				It("should have init container", func() {
					podTemplate := bookkeepercluster.MakeBookiePodTemplate(bk)
					Ω(podTemplate.Spec.InitContainers[0].Name).To(Equal("testing"))
					Ω(podTemplate.Spec.InitContainers[0].Image).To(Equal("dummy-image"))
					Ω(strings.Contains(podTemplate.Spec.InitContainers[0].Command[2], "ls;pwd")).Should(BeTrue())
				})
				It("should have security context", func() {
					podTemplate := bookkeepercluster.MakeBookiePodTemplate(bk)
					Ω(fmt.Sprintf("%v", *podTemplate.Spec.SecurityContext.RunAsUser)).To(Equal("1000"))
					Ω(fmt.Sprintf("%v", *podTemplate.Spec.SecurityContext.RunAsGroup)).To(Equal("1000"))
					Ω(fmt.Sprintf("%v", *podTemplate.Spec.SecurityContext.FSGroup)).To(Equal("1000"))
				})
			})
		})

		Context("User is not specifying bookkeeper journal and ledger path", func() {
			BeforeEach(func() {
				bk.Spec = v1alpha1.BookkeeperClusterSpec{
					Options: map[string]string{
						"hostPathVolumeMounts": "foo=/tmp/foo",
						"emptyDirVolumeMounts": "baz=/tmp/baz",
					},
				}
				bk.WithDefaults()
			})
			Context("Bookkeeper", func() {
				It("should create a headless service", func() {
					headlessService := bookkeepercluster.MakeBookieHeadlessService(bk)
					Ω(headlessService.Name).Should(Equal(bk.HeadlessServiceNameForBookie()))
				})

				It("should create a pod disruption budget", func() {
					pdb := bookkeepercluster.MakeBookiePodDisruptionBudget(bk)
					Ω(pdb.Name).Should(Equal(util.PdbNameForBookie(bk.Name)))
				})

				It("should create a config-map", func() {
					cm := bookkeepercluster.MakeBookieConfigMap(bk)
					Ω(cm.Name).Should(Equal(util.ConfigMapNameForBookie(bk.Name)))
				})

				It("should create a stateful set", func() {
					ss := bookkeepercluster.MakeBookieStatefulSet(bk)
					Ω(ss.Name).Should(Equal(util.StatefulSetNameForBookie(bk.Name)))
				})

				It("should have journal and ledgers dir set to default value", func() {
					sts := bookkeepercluster.MakeBookieStatefulSet(bk)
					mountledger := sts.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath
					Ω(mountledger).Should(Equal("/bk/ledgers"))
					mountjournal := sts.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath
					Ω(mountjournal).Should(Equal("/bk/journal"))
					indexjournal := sts.Spec.Template.Spec.Containers[0].VolumeMounts[2].MountPath
					Ω(indexjournal).Should(Equal("/bk/index"))
				})

				It("should have hostPathVolumeMounts set to the value given by user", func() {
					sts := bookkeepercluster.MakeBookieStatefulSet(bk)
					mounthostpath := sts.Spec.Template.Spec.Containers[0].VolumeMounts[3].MountPath
					Ω(mounthostpath).Should(Equal("/tmp/foo"))
				})

				It("should have emptyDirVolumeMounts set to the value given by user", func() {
					sts := bookkeepercluster.MakeBookieStatefulSet(bk)
					mounthostpath := sts.Spec.Template.Spec.Containers[0].VolumeMounts[4].MountPath
					Ω(mounthostpath).Should(Equal("/tmp/baz"))
				})

				It("should have probe timeout values set to their default value", func() {
					rp_i := bk.Spec.Probes.ReadinessProbe.InitialDelaySeconds
					rp_p := bk.Spec.Probes.ReadinessProbe.PeriodSeconds
					rp_f := bk.Spec.Probes.ReadinessProbe.FailureThreshold
					rp_s := bk.Spec.Probes.ReadinessProbe.SuccessThreshold
					rp_t := bk.Spec.Probes.ReadinessProbe.TimeoutSeconds
					Ω(rp_i).Should(Equal(int32(20)))
					Ω(rp_p).Should(Equal(int32(10)))
					Ω(rp_f).Should(Equal(int32(9)))
					Ω(rp_s).Should(Equal(int32(1)))
					Ω(rp_t).Should(Equal(int32(5)))
					lp_i := bk.Spec.Probes.LivenessProbe.InitialDelaySeconds
					lp_p := bk.Spec.Probes.LivenessProbe.PeriodSeconds
					lp_f := bk.Spec.Probes.LivenessProbe.FailureThreshold
					lp_s := bk.Spec.Probes.LivenessProbe.SuccessThreshold
					lp_t := bk.Spec.Probes.LivenessProbe.TimeoutSeconds
					Ω(lp_i).Should(Equal(int32(60)))
					Ω(lp_p).Should(Equal(int32(15)))
					Ω(lp_f).Should(Equal(int32(4)))
					Ω(lp_s).Should(Equal(int32(1)))
					Ω(lp_t).Should(Equal(int32(5)))
				})
			})
		})
	})
})
