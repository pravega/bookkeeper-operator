/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package bookkeepercluster

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	LedgerDiskName  = "ledger"
	JournalDiskName = "journal"
	IndexDiskName   = "index"
	heapDumpName    = "heap-dump"
	heapDumpDir     = "/tmp/dumpfile/heap"
)

func MakeBookieHeadlessService(bk *v1alpha1.BookkeeperCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.HeadlessServiceNameForBookie(bk.Name),
			Namespace: bk.Namespace,
			Labels:    util.LabelsForBookie(bk),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "bookie",
					Port: 3181,
				},
			},
			Selector:  util.LabelsForBookie(bk),
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

func MakeBookieStatefulSet(bk *v1alpha1.BookkeeperCluster) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.StatefulSetNameForBookie(bk.Name),
			Namespace: bk.Namespace,
			Labels:    util.LabelsForBookie(bk),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         util.HeadlessServiceNameForBookie(bk.Name),
			Replicas:            &bk.Spec.Replicas,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			Template: MakeBookiePodTemplate(bk),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForBookie(bk),
			},
			VolumeClaimTemplates: makeBookieVolumeClaimTemplates(bk),
		},
	}
}

func MakeBookiePodTemplate(bk *v1alpha1.BookkeeperCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      util.LabelsForBookie(bk),
			Annotations: map[string]string{"bookkeeper.version": bk.Spec.Version},
		},
		Spec: *makeBookiePodSpec(bk),
	}
}

func makeBookiePodSpec(bk *v1alpha1.BookkeeperCluster) *corev1.PodSpec {
	environment := []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: util.ConfigMapNameForBookie(bk.Name),
				},
			},
		},
	}

	configMapName := strings.TrimSpace(bk.Spec.EnvVars)
	if configMapName != "" {
		environment = append(environment, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		})
	}

	var ledgerDirs, journalDirs, indexDirs []string
	var ok bool

	if _, ok = bk.Spec.Options["ledgerDirectories"]; ok {
		ledgerDirs = strings.Split(bk.Spec.Options["ledgerDirectories"], ",")
	} else {
		// default value if user did not set ledgerDirectories in options
		ledgerDirs = append(ledgerDirs, "/bk/ledgers")
	}

	if _, ok = bk.Spec.Options["journalDirectories"]; ok {
		journalDirs = strings.Split(bk.Spec.Options["journalDirectories"], ",")
	} else {
		// default value if user did not set journalDirectories in options
		journalDirs = append(journalDirs, "/bk/journal")
	}

	if _, ok = bk.Spec.Options["indexDirectories"]; ok {
		indexDirs = strings.Split(bk.Spec.Options["indexDirectories"], ",")
	} else {
		// default value if user did not set indexDirectories in options
		indexDirs = append(indexDirs, "/bk/index")
	}

	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "bookie",
				Image:           util.BookkeeperImage(bk),
				ImagePullPolicy: bk.Spec.Image.PullPolicy,
				Ports: []corev1.ContainerPort{
					{
						Name:          "bookie",
						ContainerPort: 3181,
					},
				},
				EnvFrom:      environment,
				VolumeMounts: createVolumeMount(ledgerDirs, journalDirs, indexDirs),
				Resources:    *bk.Spec.Resources,
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/sh", "-c", "/opt/bookkeeper/bin/bookkeeper shell bookiesanity"},
						},
					},
					// Bookie pods should start fast. We give it up to 1.5 minute to become ready.
					InitialDelaySeconds: 20,
					PeriodSeconds:       10,
					FailureThreshold:    9,
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: util.HealthcheckCommand(3181),
						},
					},
					// We start the liveness probe from the maximum time the pod can take
					// before becoming ready.
					// If the pod fails the health check during 1 minute, Kubernetes
					// will restart it.
					InitialDelaySeconds: 60,
					PeriodSeconds:       15,
					FailureThreshold:    4,
				},
			},
		},
		Affinity: util.PodAntiAffinity("bookie", bk.Name),
		Volumes: []corev1.Volume{
			{
				Name: heapDumpName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	if bk.Spec.ServiceAccountName != "" {
		podSpec.ServiceAccountName = bk.Spec.ServiceAccountName
	}

	return podSpec
}

func createVolumeMount(ledgerDirs []string, journalDirs []string, indexDirs []string) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount
	for i, ledger := range ledgerDirs {
		name := LedgerDiskName + strconv.Itoa(i)
		v := corev1.VolumeMount{
			Name:      LedgerDiskName,
			MountPath: ledger,
			SubPath:   name,
		}
		volumeMounts = append(volumeMounts, v)
	}
	for i, journal := range journalDirs {
		name := JournalDiskName + strconv.Itoa(i)
		v := corev1.VolumeMount{
			Name:      JournalDiskName,
			MountPath: journal,
			SubPath:   name,
		}
		volumeMounts = append(volumeMounts, v)
	}
	for i, index := range indexDirs {
		name := IndexDiskName + strconv.Itoa(i)
		v := corev1.VolumeMount{
			Name:      IndexDiskName,
			MountPath: index,
			SubPath:   name,
		}
		volumeMounts = append(volumeMounts, v)
	}
	return volumeMounts
}

func makeBookieVolumeClaimTemplates(bk *v1alpha1.BookkeeperCluster) []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      JournalDiskName,
				Namespace: bk.Namespace,
			},
			Spec: *bk.Spec.Storage.JournalVolumeClaimTemplate,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      LedgerDiskName,
				Namespace: bk.Namespace,
			},
			Spec: *bk.Spec.Storage.LedgerVolumeClaimTemplate,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      IndexDiskName,
				Namespace: bk.Namespace,
			},
			Spec: *bk.Spec.Storage.IndexVolumeClaimTemplate,
		},
	}
}

func MakeBookieConfigMap(bk *v1alpha1.BookkeeperCluster) *corev1.ConfigMap {
	memoryOpts := []string{
		"-Xms1g",
		"-XX:MaxDirectMemorySize=1g",
		"-XX:+ExitOnOutOfMemoryError",
		"-XX:+CrashOnOutOfMemoryError",
		"-XX:+HeapDumpOnOutOfMemoryError",
		"-XX:HeapDumpPath=" + heapDumpDir,
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:+UseCGroupMemoryLimitForHeap",
		"-XX:MaxRAMFraction=2",
	}

	memoryOpts = util.OverrideDefaultJVMOptions(memoryOpts, bk.Spec.JVMOptions.MemoryOpts)

	gcOpts := []string{
		"-XX:+UseG1GC",
		"-XX:MaxGCPauseMillis=10",
		"-XX:+ParallelRefProcEnabled",
		"-XX:+AggressiveOpts",
		"-XX:+DoEscapeAnalysis",
		"-XX:ParallelGCThreads=32",
		"-XX:ConcGCThreads=32",
		"-XX:G1NewSizePercent=50",
		"-XX:+DisableExplicitGC",
		"-XX:-ResizePLAB",
	}
	gcOpts = util.OverrideDefaultJVMOptions(gcOpts, bk.Spec.JVMOptions.GcOpts)

	gcLoggingOpts := []string{
		"-XX:+PrintGCDetails",
		"-XX:+PrintGCDateStamps",
		"-XX:+PrintGCApplicationStoppedTime",
		"-XX:+UseGCLogFileRotation",
		"-XX:NumberOfGCLogFiles=5",
		"-XX:GCLogFileSize=64m",
	}
	gcLoggingOpts = util.OverrideDefaultJVMOptions(gcLoggingOpts, bk.Spec.JVMOptions.GcLoggingOpts)

	extraOpts := []string{}
	if bk.Spec.JVMOptions.ExtraOpts != nil {
		extraOpts = bk.Spec.JVMOptions.ExtraOpts
	}

	configData := map[string]string{
		"BOOKIE_MEM_OPTS":          strings.Join(memoryOpts, " "),
		"BOOKIE_GC_OPTS":           strings.Join(gcOpts, " "),
		"BOOKIE_GC_LOGGING_OPTS":   strings.Join(gcLoggingOpts, " "),
		"BOOKIE_EXTRA_OPTS":        strings.Join(extraOpts, " "),
		"ZK_URL":                   bk.Spec.ZookeeperUri,
		"BK_useHostNameAsBookieID": "true",
	}

	if match, _ := util.CompareVersions(bk.Spec.Version, "0.5.0", "<"); match {
		// bookkeeper < 0.5 uses BookKeeper 4.5, which does not play well
		// with hostnames that resolve to different IP addresses over time
		configData["BK_useHostNameAsBookieID"] = "false"
	}

	if *bk.Spec.AutoRecovery {
		configData["BK_autoRecoveryDaemonEnabled"] = "true"
		// Wait one minute before starting autorecovery. This will give
		// pods some time to come up after being updated or migrated
		configData["BK_lostBookieRecoveryDelay"] = "60"
	} else {
		configData["BK_autoRecoveryDaemonEnabled"] = "false"
	}

	for k, v := range bk.Spec.Options {
		prefixKey := fmt.Sprintf("BK_%s", k)
		configData[prefixKey] = v
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ConfigMapNameForBookie(bk.Name),
			Namespace: bk.ObjectMeta.Namespace,
		},
		Data: configData,
	}
}

func MakeBookiePodDisruptionBudget(bk *v1alpha1.BookkeeperCluster) *policyv1beta1.PodDisruptionBudget {
	maxUnavailable := intstr.FromInt(1)
	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.PdbNameForBookie(bk.Name),
			Namespace: bk.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForBookie(bk),
			},
		},
	}
}
