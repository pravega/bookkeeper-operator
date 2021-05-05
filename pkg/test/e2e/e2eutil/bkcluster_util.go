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
	goctx "context"
	"fmt"
	"strings"
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	bkapi "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	RetryInterval        = time.Second * 5
	Timeout              = time.Second * 60
	CleanupRetryInterval = time.Second * 1
	CleanupTimeout       = time.Second * 5
	ReadyTimeout         = time.Minute * 5
	UpgradeTimeout       = time.Minute * 10
	TerminateTimeout     = time.Minute * 2
	VerificationTimeout  = time.Minute * 5
)

// CreateBKCluster creates a BookkeeperCluster CR with the desired spec
func CreateBKCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster) (*bkapi.BookkeeperCluster, error) {
	t.Logf("creating bookkeeper cluster: %s", b.Name)
	b.Spec.EnvVars = "bookkeeper-configmap"
	b.Spec.ZookeeperUri = "zookeeper-client:2181"
	b.Spec.Image.ImageSpec.PullPolicy = "IfNotPresent"
	b.Spec.Probes.ReadinessProbe.PeriodSeconds = 15
	b.Spec.Probes.ReadinessProbe.TimeoutSeconds = 10
	b.Spec.Storage.LedgerVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
	}
	b.Spec.Storage.IndexVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
	}
	b.Spec.Storage.JournalVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
	}
	err := f.Client.Create(goctx.TODO(), b, &framework.CleanupOptions{TestContext: ctx, Timeout: CleanupTimeout, RetryInterval: CleanupRetryInterval})
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}

	bookkeeper := &bkapi.BookkeeperCluster{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, bookkeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	t.Logf("created bookkeeper cluster: %s", b.Name)
	return bookkeeper, nil
}

// CreateBKCluster creates a BookkeeperCluster CR with the desired spec
func CreateBKClusterWithCM(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster, cm string) (*bkapi.BookkeeperCluster, error) {
	t.Logf("creating bookkeeper cluster: %s", b.Name)
	b.Spec.EnvVars = cm
	b.Spec.ZookeeperUri = "zookeeper-client:2181"
	b.Spec.Image.ImageSpec.PullPolicy = "IfNotPresent"
	b.Spec.Probes.ReadinessProbe.PeriodSeconds = 15
	b.Spec.Probes.ReadinessProbe.TimeoutSeconds = 10
	b.Spec.Storage.LedgerVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
	}
	b.Spec.Storage.IndexVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
	}
	b.Spec.Storage.JournalVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
	}
	err := f.Client.Create(goctx.TODO(), b, &framework.CleanupOptions{TestContext: ctx, Timeout: CleanupTimeout, RetryInterval: CleanupRetryInterval})
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}

	bookkeeper := &bkapi.BookkeeperCluster{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, bookkeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	t.Logf("created bookkeeper cluster: %s", b.Name)
	return bookkeeper, nil
}

// CreateConfigMap creates the configmap specified
func CreateConfigMap(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, cm *corev1.ConfigMap) error {
	err := f.Client.Create(goctx.TODO(), cm, &framework.CleanupOptions{TestContext: ctx, Timeout: CleanupTimeout, RetryInterval: CleanupRetryInterval})
	if err != nil {
		return fmt.Errorf("failed to create Configmap: %v", err)
	}
	t.Logf("created configmap: %s", cm.ObjectMeta.Name)
	return nil
}

func DeletePods(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster, size int) error {
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"bookkeeper_cluster": b.GetName()}).String(),
	}
	podList, err := f.KubeClient.CoreV1().Pods(b.Namespace).List(listOptions)
	if err != nil {
		return err
	}
	pod := &corev1.Pod{}

	for i := 0; i < size; i++ {
		pod = &podList.Items[i]
		t.Logf("pod name is %v", pod.Name)
		err := f.Client.Delete(goctx.TODO(), pod)
		if err != nil {
			return fmt.Errorf("failed to delete pod: %v", err)
		}
		t.Logf("deleted bookkeeper pod: %s", pod.Name)
	}
	return nil
}

// DeleteBKCluster deletes the BookkeeperCluster CR specified by cluster spec
func DeleteBKCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster) error {
	t.Logf("deleting bookkeeper cluster: %s", b.Name)
	err := f.Client.Delete(goctx.TODO(), b)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete CR: %v", err)
	}

	t.Logf("deleted bookkeeper cluster: %s", b.Name)
	return nil
}

// DeleteConfigMap deletes the configmap specified
func DeleteConfigMap(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, cm *corev1.ConfigMap) error {
	t.Logf("deleting configmap: %s", cm.ObjectMeta.Name)
	err := f.Client.Delete(goctx.TODO(), cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete CM: %v", err)
	}

	t.Logf("deleted configmap: %s", cm.ObjectMeta.Name)
	return nil
}

// UpdateBkCluster updates the BookkeeperCluster CR
func UpdateBKCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster) error {
	t.Logf("updating bookkeeper cluster: %s", b.Name)
	err := f.Client.Update(goctx.TODO(), b)
	if err != nil {
		return fmt.Errorf("failed to update CR: %v", err)
	}

	t.Logf("updated bookkeeper cluster: %s", b.Name)
	return nil
}

// GetBKCluster returns the latest BookkeeperCluster CR
func GetBKCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster) (*bkapi.BookkeeperCluster, error) {
	bookkeeper := &bkapi.BookkeeperCluster{}
	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, bookkeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	return bookkeeper, nil
}

func CheckEvents(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster, event string) (bool, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"bookkeeper_cluster": b.GetName()}).String(),
	}

	events, err := f.KubeClient.CoreV1().Events(b.Namespace).List(listOptions)
	if err != nil {
		return false, err
	}

	if events != nil {
		for _, e := range events.Items {
			if strings.HasPrefix(e.Name, event) {
				return true, nil
			}
		}
	}

	return false, nil
}

func CheckConfigMap(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster, key string, value string) error {
	cm := &corev1.ConfigMap{}
	name := util.ConfigMapNameForBookie(b.Name)
	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: b.Namespace, Name: name}, cm)
	if err != nil {
		return fmt.Errorf("failed to obtain configmap: %v", err)
	}
	if cm != nil {
		if cm.Data[key] == value {
			return nil
		}
	}
	return fmt.Errorf("Configmap does not contain the expected value")
}

// WaitForBookkeeperClusterToBecomeReady will wait until all Bookkeeper cluster pods are ready
func WaitForBookkeeperClusterToBecomeReady(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster) error {
	t.Logf("waiting for cluster pods to become ready: %s", b.Name)

	err := wait.Poll(RetryInterval, ReadyTimeout, func() (done bool, err error) {
		cluster, err := GetBKCluster(t, f, ctx, b)

		if err != nil {
			return false, err
		}

		t.Logf("\twaiting for pods to become ready (%d/%d), pods (%v)", cluster.Status.ReadyReplicas, cluster.Spec.Replicas, cluster.Status.Members.Ready)

		_, condition := cluster.Status.GetClusterCondition(bkapi.ClusterConditionPodsReady)
		if condition != nil && condition.Status == corev1.ConditionTrue && cluster.Status.ReadyReplicas == cluster.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	t.Logf("bookkeeper cluster ready: %s", b.Name)
	return nil
}

// WaitForBKClusterToTerminate will wait until all Bookkeeper cluster pods are terminated
func WaitForBKClusterToTerminate(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster) error {
	t.Logf("waiting for Bookkeeper cluster to terminate: %s", b.Name)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"bookkeeper_cluster": b.GetName()}).String(),
	}

	// Wait for Pods to terminate
	err := wait.Poll(RetryInterval, TerminateTimeout, func() (done bool, err error) {
		podList, err := f.KubeClient.CoreV1().Pods(b.Namespace).List(listOptions)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}
		t.Logf("waiting for pods to terminate, running pods (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	// Wait for PVCs to terminate
	err = wait.Poll(RetryInterval, TerminateTimeout, func() (done bool, err error) {
		pvcList, err := f.KubeClient.CoreV1().PersistentVolumeClaims(b.Namespace).List(listOptions)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range pvcList.Items {
			pvc := &pvcList.Items[i]
			names = append(names, pvc.Name)
		}
		t.Logf("waiting for pvc to terminate (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil

	})

	if err != nil {
		return err
	}

	t.Logf("bookkeeper cluster terminated: %s", b.Name)
	return nil
}

// WaitForBookkeeperClusterToUpgrade will wait until all pods are upgraded
func WaitForBKClusterToUpgrade(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster, targetVersion string) error {
	t.Logf("waiting for cluster to upgrade: %s", b.Name)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"bookkeeper_cluster": b.GetName()}).String(),
	}

	err := wait.Poll(RetryInterval, UpgradeTimeout, func() (done bool, err error) {
		cluster, err := GetBKCluster(t, f, ctx, b)
		if err != nil {
			return false, err
		}

		_, upgradeCondition := cluster.Status.GetClusterCondition(bkapi.ClusterConditionUpgrading)
		_, errorCondition := cluster.Status.GetClusterCondition(bkapi.ClusterConditionError)

		t.Logf("\twaiting for cluster to upgrade (upgrading: %s; error: %s)", upgradeCondition.Status, errorCondition.Status)

		if errorCondition.Status == corev1.ConditionTrue && errorCondition.Reason == "UpgradeFailed" {
			return false, fmt.Errorf("failed upgrading cluster: [%s] %s", errorCondition.Reason, errorCondition.Message)
		}

		if upgradeCondition.Status == corev1.ConditionFalse && cluster.Status.CurrentVersion == targetVersion {
			// Cluster upgraded
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	// check whether PVCs have been reattached
	pvcList, err := f.KubeClient.CoreV1().PersistentVolumeClaims(b.Namespace).List(listOptions)
	if err != nil {
		return err
	}

	index, journal, ledger := int32(0), int32(0), int32(0)

	for i := range pvcList.Items {
		pvc := &pvcList.Items[i]
		if strings.HasPrefix(pvc.Name, "index") {
			index++
		} else if strings.HasPrefix(pvc.Name, "journal") {
			journal++
		} else if strings.HasPrefix(pvc.Name, "ledger") {
			ledger++
		}
	}

	if index != b.Spec.Replicas || journal != b.Spec.Replicas || ledger != b.Spec.Replicas {
		return fmt.Errorf("PVC count mismatch")
	}

	t.Logf("bookkeeper cluster upgraded: %s", b.Name)
	return nil
}

func WaitForCMBKClusterToUpgrade(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster) error {
	t.Logf("waiting for cluster to upgrade post cm changes: %s", b.Name)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"bookkeeper_cluster": b.GetName()}).String(),
	}

	// Checking if all pods are getting restarted
	podList, err := f.KubeClient.CoreV1().Pods(b.Namespace).List(listOptions)
	if err != nil {
		return err
	}

	for i := range podList.Items {
		pod := &podList.Items[i]
		name := pod.Name
		t.Logf("waiting for pods to terminate, running pods (%v)", pod.Name)
		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: b.Namespace, Name: name}, pod)
		start := time.Now()
		for util.IsPodReady(pod) {
			if time.Since(start) > 5*time.Minute {
				return fmt.Errorf("failed to delete Bookkeeper pod (%s) for 5 mins ", name)
			}
			err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: b.Namespace, Name: name}, pod)
		}
	}

	err = wait.Poll(RetryInterval, ReadyTimeout, func() (done bool, err error) {
		cluster, err := GetBKCluster(t, f, ctx, b)

		if err != nil {
			return false, err
		}

		t.Logf("\twaiting for pods to become ready (%d/%d), pods (%v)", cluster.Status.ReadyReplicas, cluster.Spec.Replicas, cluster.Status.Members.Ready)

		_, condition := cluster.Status.GetClusterCondition(bkapi.ClusterConditionPodsReady)
		if condition != nil && condition.Status == corev1.ConditionTrue && cluster.Status.ReadyReplicas == cluster.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})

	// check whether PVCs have been reattached
	pvcList, err := f.KubeClient.CoreV1().PersistentVolumeClaims(b.Namespace).List(listOptions)
	if err != nil {
		return err
	}

	index, journal, ledger := int32(0), int32(0), int32(0)

	for i := range pvcList.Items {
		pvc := &pvcList.Items[i]
		if strings.HasPrefix(pvc.Name, "index") {
			index++
		} else if strings.HasPrefix(pvc.Name, "journal") {
			journal++
		} else if strings.HasPrefix(pvc.Name, "ledger") {
			ledger++
		}
	}

	if index != b.Spec.Replicas || journal != b.Spec.Replicas || ledger != b.Spec.Replicas {
		return fmt.Errorf("PVC count mismatch")
	}

	t.Logf("bookkeeper cluster updated: %s", b.Name)
	return nil
}

// WaitForBookkeeperClusterToRollback will wait until all pods are rolled back
func WaitForBKClusterToRollback(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, b *bkapi.BookkeeperCluster, targetVersion string) error {
	t.Logf("waiting for cluster to rollback: %s", b.Name)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"bookkeeper_cluster": b.GetName()}).String(),
	}

	err := wait.Poll(RetryInterval, UpgradeTimeout, func() (done bool, err error) {
		cluster, err := GetBKCluster(t, f, ctx, b)
		if err != nil {
			return false, err
		}

		_, rollbackCondition := cluster.Status.GetClusterCondition(bkapi.ClusterConditionRollback)
		_, errorCondition := cluster.Status.GetClusterCondition(bkapi.ClusterConditionError)

		t.Logf("\twaiting for cluster to rollback (rollback in progress: %s)", rollbackCondition.Status)

		if errorCondition.Status == corev1.ConditionTrue && errorCondition.Reason == "RollbackFailed" {
			return false, fmt.Errorf("failed rolling back cluster: [%s] %s", errorCondition.Reason, errorCondition.Message)
		}

		if rollbackCondition.Status == corev1.ConditionFalse && cluster.Status.CurrentVersion == targetVersion {
			// Cluster rolled back
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	// check whether PVCs have been reattached
	pvcList, err := f.KubeClient.CoreV1().PersistentVolumeClaims(b.Namespace).List(listOptions)
	if err != nil {
		return err
	}

	index, journal, ledger := int32(0), int32(0), int32(0)

	for i := range pvcList.Items {
		pvc := &pvcList.Items[i]
		if strings.HasPrefix(pvc.Name, "index") {
			index++
		} else if strings.HasPrefix(pvc.Name, "journal") {
			journal++
		} else if strings.HasPrefix(pvc.Name, "ledger") {
			ledger++
		}
	}

	if index != b.Spec.Replicas || journal != b.Spec.Replicas || ledger != b.Spec.Replicas {
		return fmt.Errorf("PVC count mismatch")
	}

	t.Logf("bookkeeper cluster rolled back: %s", b.Name)
	return nil
}
