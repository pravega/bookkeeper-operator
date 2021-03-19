/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1alpha1

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	k8s "github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/pravega/bookkeeper-operator/pkg/controller/config"
	"github.com/pravega/bookkeeper-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var Mgr manager.Manager

const (
	// DefaultZookeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultZookeeperUri = "zookeeper-client:2181"

	// DefaultBookkeeperVersion is the default tag used for for the BookKeeper
	// Docker image
	DefaultBookkeeperVersion = "0.7.0"
	// DefaultBookkeeperImageRepository is the default Docker repository for

	// the BookKeeper image
	DefaultBookkeeperImageRepository = "pravega/bookkeeper"

	// DefaultbookkeeperImagePullPolicy is the default image pull policy used
	// for the Bookkeeper Docker image
	DefaultBookkeeperImagePullPolicy = corev1.PullAlways

	// DefaultBookkeeperLedgerVolumeSize is the default volume size for the
	// Bookkeeper ledger volume
	DefaultBookkeeperLedgerVolumeSize = "10Gi"

	// DefaultBookkeeperJournalVolumeSize is the default volume size for the
	// Bookkeeper journal volume
	DefaultBookkeeperJournalVolumeSize = "10Gi"

	// DefaultBookkeeperIndexVolumeSize is the default volume size for the
	// Bookkeeper index volume
	DefaultBookkeeperIndexVolumeSize = "10Gi"

	// MinimumBookkeeperReplicas is the minimum number of Bookkeeper replicas
	// accepted
	MinimumBookkeeperReplicas = 3

	// DefaultBookkeeperRequestCPU is the default CPU request for BookKeeper
	DefaultBookkeeperRequestCPU = "500m"

	// DefaultBookkeeperLimitCPU is the default CPU limit for BookKeeper
	DefaultBookkeeperLimitCPU = "1"

	// DefaultBookkeeperRequestMemory is the default memory request for BookKeeper
	DefaultBookkeeperRequestMemory = "1Gi"

	// DefaultBookkeeperLimitMemory is the limit memory limit for BookKeeper
	DefaultBookkeeperLimitMemory = "2Gi"

	// DefaultReadinessProbeInitialDelaySeconds is the default initial delay (in seconds)
	// for the readiness probe
	DefaultReadinessProbeInitialDelaySeconds = 20

	// DefaultReadinessProbePeriodSeconds is the default probe period (in seconds)
	// for the readiness probe
	DefaultReadinessProbePeriodSeconds = 10

	// DefaultReadinessProbeFailureThreshold is the default probe failure threshold
	// for the readiness probe
	DefaultReadinessProbeFailureThreshold = 9

	// DefaultReadinessProbeSuccessThreshold is the default probe success threshold
	// for the readiness probe
	DefaultReadinessProbeSuccessThreshold = 1

	// DefaultReadinessProbeTimeoutSeconds is the default probe timeout (in seconds)
	// for the readiness probe
	DefaultReadinessProbeTimeoutSeconds = 5

	// DefaultLivenessProbeInitialDelaySeconds is the default initial delay (in seconds)
	// for the liveness probe
	DefaultLivenessProbeInitialDelaySeconds = 60

	// DefaultLivenessProbePeriodSeconds is the default probe period (in seconds)
	// for the liveness probe
	DefaultLivenessProbePeriodSeconds = 15

	// DefaultLivenessProbeFailureThreshold is the default probe failure threshold
	// for the liveness probe
	DefaultLivenessProbeFailureThreshold = 4

	// DefaultLivenessProbeSuccessThreshold is the default probe success threshold
	// for the liveness probe
	DefaultLivenessProbeSuccessThreshold = 1

	// DefaultLivenessProbeTimeoutSeconds is the default probe timeout (in seconds)
	// for the liveness probe
	DefaultLivenessProbeTimeoutSeconds = 5
)

func init() {
	SchemeBuilder.Register(&BookkeeperCluster{}, &BookkeeperClusterList{})
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BookkeeperClusterList contains a list of BookkeeperCluster
type BookkeeperClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BookkeeperCluster `json:"items"`
}

// Generate CRD using kubebuilder
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=bk
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.currentVersion`,description="The current bookkeeper version"
// +kubebuilder:printcolumn:name="Desired Version",type=string,JSONPath=`.spec.version`,description="The desired bookkeeper version"
// +kubebuilder:printcolumn:name="Desired Members",type=integer,JSONPath=`.status.replicas`,description="The number of desired bookkeeper members"
// +kubebuilder:printcolumn:name="Ready Members",type=integer,JSONPath=`.status.readyReplicas`,description="The number of ready bookkeeper members"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// BookkeeperCluster is the Schema for the BookkeeperClusters API
type BookkeeperCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookkeeperClusterSpec   `json:"spec,omitempty"`
	Status BookkeeperClusterStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (bk *BookkeeperCluster) WithDefaults() (changed bool) {
	changed = bk.Spec.withDefaults()
	return changed
}

// ClusterSpec defines the desired state of BookkeeperCluster
type BookkeeperClusterSpec struct {
	// ZookeeperUri specifies the hostname/IP address and port in the format
	// "hostname:port".
	// By default, the value "zookeeper-client:2181" is used, that corresponds to the
	// default Zookeeper service created by the Pravega Zookkeeper operator
	// available at: https://github.com/pravega/zookeeper-operator
	// +optional
	ZookeeperUri string `json:"zookeeperUri"`

	// Image defines the BookKeeper Docker image to use.
	// By default, "pravega/bookkeeper" will be used.
	// +optional
	Image *BookkeeperImageSpec `json:"image"`

	// Replicas defines the number of BookKeeper replicas.
	// Minimum is 3. Defaults to 3.
	// If testmode is enabled, 1 replica is allowed.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas int32 `json:"replicas"`

	// MaxUnavailableBookkeeperReplicas defines the
	// MaxUnavailable Bookkeeper Replicas
	// Default is 1.
	// +optional
	MaxUnavailableBookkeeperReplicas int32 `json:"maxUnavailableBookkeeperReplicas"`

	// Storage configures the storage for BookKeeper
	// +optional
	Storage *BookkeeperStorageSpec `json:"storage"`

	// AutoRecovery indicates whether or not BookKeeper auto recovery is enabled.
	// Defaults to true.
	// +optional
	AutoRecovery *bool `json:"autoRecovery"`

	// ServiceAccountName configures the service account used on BookKeeper instances
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Probes specifies the timeout values for the Readiness and Liveness Probes
	// for the bookkeeper pods.
	// +optional
	Probes *Probes `json:"probes"`

	// BookieResources specifies the request and limit of resources that bookie can have.
	// BookieResources includes CPU and memory resources
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Options is the Bookkeeper configuration that is to override the bk_server.conf
	// in bookkeeper. Some examples can be found here
	// https://github.com/apache/bookkeeper/blob/master/docker/README.md
	// +optional
	Options map[string]string `json:"options"`

	// JVM is the JVM options for bookkeeper. It will be passed to the JVM for performance tuning.
	// If this field is not specified, the operator will use a set of default
	// options that is good enough for general deployment.
	// +optional
	JVMOptions *JVMOptions `json:"jvmOptions"`

	// Provides the name of the configmap created by the user to provide additional key-value pairs
	// that need to be configured into the bookie pods as environmental variables
	EnvVars string `json:"envVars,omitempty"`

	// Version is the expected version of the Bookkeeper cluster.
	// The bookkeeper-operator will eventually make the Bookkeeper cluster version
	// equal to the expected version.
	//
	// The version must follow the [semver]( http://semver.org) format, for example "3.2.13".
	// Only Bookkeeper released versions are supported: https://hub.docker.com/r/pravega/bookkeeper/tags
	//
	// If version is not set, default is "0.4.0".
	// +optional
	Version string `json:"version"`
	// If true, AND if the owner has the "foregroundDeletion" finalizer, then
	// the owner cannot be deleted from the key-value store until this
	// reference is removed.
	// Defaults to true
	BlockOwnerDeletion *bool `json:"blockOwnerDeletion,omitempty"`

	// Labels to be added to the bookie pods
	// +optional
	Labels map[string]string `json:"labels"`
}

// BookkeeperImageSpec defines the fields needed for a BookKeeper Docker image
type BookkeeperImageSpec struct {
	ImageSpec `json:"imageSpec,omitempty"`
}

func (s *BookkeeperImageSpec) withDefaults() (changed bool) {
	if s.Repository == "" {
		changed = true
		s.Repository = DefaultBookkeeperImageRepository
	}

	s.Tag = ""

	if s.PullPolicy == "" {
		changed = true
		s.PullPolicy = DefaultBookkeeperImagePullPolicy
	}

	return changed
}

type Probes struct {
	// +optional
	ReadinessProbe *Probe `json:"readinessProbe"`
	// +optional
	LivenessProbe *Probe `json:"livenessProbe"`
}

func (s *Probes) withDefaults() (changed bool) {
	if s.ReadinessProbe == nil {
		changed = true
		s.ReadinessProbe = &Probe{}
		s.ReadinessProbe.InitialDelaySeconds = DefaultReadinessProbeInitialDelaySeconds
		s.ReadinessProbe.PeriodSeconds = DefaultReadinessProbePeriodSeconds
		s.ReadinessProbe.FailureThreshold = DefaultReadinessProbeFailureThreshold
		s.ReadinessProbe.SuccessThreshold = DefaultReadinessProbeSuccessThreshold
		s.ReadinessProbe.TimeoutSeconds = DefaultReadinessProbeTimeoutSeconds
	}

	if s.LivenessProbe == nil {
		changed = true
		s.LivenessProbe = &Probe{}
		s.LivenessProbe.InitialDelaySeconds = DefaultLivenessProbeInitialDelaySeconds
		s.LivenessProbe.PeriodSeconds = DefaultLivenessProbePeriodSeconds
		s.LivenessProbe.FailureThreshold = DefaultLivenessProbeFailureThreshold
		s.LivenessProbe.SuccessThreshold = DefaultLivenessProbeSuccessThreshold
		s.LivenessProbe.TimeoutSeconds = DefaultLivenessProbeTimeoutSeconds
	}

	return changed
}

type Probe struct {
	// +kubebuilder:validation:Minimum=0
	// +optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	PeriodSeconds int32 `json:"periodSeconds"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	FailureThreshold int32 `json:"failureThreshold"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	SuccessThreshold int32 `json:"successThreshold"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds"`
}

type JVMOptions struct {
	// +optional
	MemoryOpts []string `json:"memoryOpts"`
	// +optional
	GcOpts []string `json:"gcOpts"`
	// +optional
	GcLoggingOpts []string `json:"gcLoggingOpts"`
	// +optional
	ExtraOpts []string `json:"extraOpts"`
}

func (s *JVMOptions) withDefaults() (changed bool) {
	if s.MemoryOpts == nil {
		changed = true
		s.MemoryOpts = []string{}
	}

	if s.GcOpts == nil {
		changed = true
		s.GcOpts = []string{}
	}

	if s.GcLoggingOpts == nil {
		changed = true
		s.GcLoggingOpts = []string{}
	}

	if s.ExtraOpts == nil {
		changed = true
		s.ExtraOpts = []string{}
	}

	return changed
}

// BookkeeperStorageSpec is the configuration of the volumes used in BookKeeper
type BookkeeperStorageSpec struct {
	// LedgerVolumeClaimTemplate is the spec to describe PVC for the BookKeeper ledger
	// This field is optional. If no PVC spec and there is no default storage class,
	// stateful containers will use emptyDir as volume
	// +optional
	LedgerVolumeClaimTemplate *corev1.PersistentVolumeClaimSpec `json:"ledgerVolumeClaimTemplate"`

	// JournalVolumeClaimTemplate is the spec to describe PVC for the BookKeeper journal
	// This field is optional. If no PVC spec and there is no default storage class,
	// stateful containers will use emptyDir as volume
	// +optional
	JournalVolumeClaimTemplate *corev1.PersistentVolumeClaimSpec `json:"journalVolumeClaimTemplate"`

	// IndexVolumeClaimTemplate is the spec to describe PVC for the BookKeeper index
	// This field is optional. If no PVC spec and there is no default storage class,
	// stateful containers will use emptyDir as volume
	// +optional
	IndexVolumeClaimTemplate *corev1.PersistentVolumeClaimSpec `json:"indexVolumeClaimTemplate"`
}

func (s *BookkeeperStorageSpec) withDefaults() (changed bool) {
	if s.LedgerVolumeClaimTemplate == nil {
		changed = true
		s.LedgerVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(DefaultBookkeeperLedgerVolumeSize),
				},
			},
		}
	}

	if s.JournalVolumeClaimTemplate == nil {
		changed = true
		s.JournalVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(DefaultBookkeeperJournalVolumeSize),
				},
			},
		}
	}

	if s.IndexVolumeClaimTemplate == nil {
		changed = true
		s.IndexVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(DefaultBookkeeperIndexVolumeSize),
				},
			},
		}
	}
	return changed
}

func (s *BookkeeperClusterSpec) withDefaults() (changed bool) {
	if s.ZookeeperUri == "" {
		changed = true
		s.ZookeeperUri = DefaultZookeeperUri
	}

	if s.Image == nil {
		changed = true
		s.Image = &BookkeeperImageSpec{}
	}
	if s.Image.withDefaults() {
		changed = true
	}

	if !config.TestMode && s.Replicas < MinimumBookkeeperReplicas {
		changed = true
		s.Replicas = MinimumBookkeeperReplicas
	}

	if !config.TestMode && s.MaxUnavailableBookkeeperReplicas < 1 {
		changed = true
		s.MaxUnavailableBookkeeperReplicas = 1
	}

	if s.Storage == nil {
		changed = true
		s.Storage = &BookkeeperStorageSpec{}
	}
	if s.Storage.withDefaults() {
		changed = true
	}

	if s.AutoRecovery == nil {
		changed = true
		boolTrue := true
		s.AutoRecovery = &boolTrue
	}

	if s.Probes == nil {
		changed = true
		s.Probes = &Probes{}
	}
	if s.Probes.withDefaults() {
		changed = true
	}

	if s.Resources == nil {
		changed = true
		s.Resources = &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(DefaultBookkeeperRequestCPU),
				corev1.ResourceMemory: resource.MustParse(DefaultBookkeeperRequestMemory),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(DefaultBookkeeperLimitCPU),
				corev1.ResourceMemory: resource.MustParse(DefaultBookkeeperLimitMemory),
			},
		}
	}

	if s.Options == nil {
		s.Options = map[string]string{}
	}

	if s.JVMOptions == nil {
		changed = true
		s.JVMOptions = &JVMOptions{}
	}

	if s.JVMOptions.withDefaults() {
		changed = true
	}

	if s.Version == "" {
		s.Version = DefaultBookkeeperVersion
		changed = true
	}

	if s.BlockOwnerDeletion == nil {
		changed = true
		boolTrue := true
		s.BlockOwnerDeletion = &boolTrue
	}

	if s.Labels == nil {
		s.Labels = map[string]string{}
	}

	return changed
}

// ImageSpec defines the fields needed for a Docker repository image
type ImageSpec struct {
	Repository string `json:"repository"`

	// Deprecated: Use `spec.Version` instead
	Tag string `json:"tag,omitempty"`
	// +kubebuilder:validation:Enum="Always";"Never";"IfNotPresent"
	PullPolicy corev1.PullPolicy `json:"pullPolicy"`
}

var _ webhook.Validator = &BookkeeperCluster{}

func (bk *BookkeeperCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	log.Print("Registering Webhook")
	return ctrl.NewWebhookManagedBy(mgr).
		For(&BookkeeperCluster{}).
		Complete()
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (bk *BookkeeperCluster) ValidateCreate() error {
	log.Printf("validate create %s", bk.Name)
	return bk.ValidateBookkeeperVersion("")
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (bk *BookkeeperCluster) ValidateUpdate(old runtime.Object) error {
	log.Printf("validate update %s", bk.Name)
	err := bk.ValidateBookkeeperVersion("")
	if err != nil {
		return err
	}
	err = bk.validateConfigMap()
	if err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (bk *BookkeeperCluster) ValidateDelete() error {
	log.Printf("validate delete %s", bk.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
func getSupportedVersions(filename string) (map[string]string, error) {

	supportedVersions := make(map[string]string)
	filepath := filename
	if filename == "" {
		filepath = "/tmp/config/keys"
	}

	file, err := os.Open(filepath)
	if err != nil {
		return supportedVersions, fmt.Errorf("Version map /tmp/config/keys not found")
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), ":")
		supportedVersions[data[0]] = data[1]
	}
	file.Close()
	return supportedVersions, nil
}
func (bk *BookkeeperCluster) ValidateBookkeeperVersion(filename string) error {
	supportedVersions, err := getSupportedVersions(filename)
	if err != nil {
		return fmt.Errorf("Error retrieving suported versions %v", err)
	}

	if bk.Spec.Version == "" {
		bk.Spec.Version = DefaultBookkeeperVersion
	}
	requestVersion := bk.Spec.Version

	if bk.Status.IsClusterInUpgradingState() && requestVersion != bk.Status.TargetVersion {
		return fmt.Errorf("failed to process the request, cluster is upgrading")
	}

	if bk.Status.IsClusterInRollbackState() {
		if requestVersion != bk.Status.GetLastVersion() {
			return fmt.Errorf("failed to process the request, rollback in progress.")
		}
	}
	if bk.Status.IsClusterInUpgradeFailedState() {
		if requestVersion != bk.Status.GetLastVersion() {
			return fmt.Errorf("Rollback to version %s not supported. Only rollback to version %s is supported.", requestVersion, bk.Status.GetLastVersion())
		}
		return nil
	}

	if bk.Status.IsClusterInErrorState() {
		return fmt.Errorf("failed to process the request, cluster is in error state.")
	}
	// Check if the request has a valid Bookkeeper version
	normRequestVersion, err := util.NormalizeVersion(requestVersion)
	log.Printf("validateBookkeeperVersion:: normRequestVersion %s", normRequestVersion)
	if err != nil {
		return fmt.Errorf("request version is not in valid format: %v", err)
	}

	if _, ok := supportedVersions[normRequestVersion]; !ok {
		return fmt.Errorf("unsupported Bookkeeper cluster version %s", requestVersion)
	}

	if bk.Status.CurrentVersion == "" {
		// we're deploying for the very first time
		return nil
	}

	// This is not an upgrade if CurrentVersion == requestVersion
	if bk.Status.CurrentVersion == requestVersion {
		return nil
	}
	// This is an upgrade, check if requested version is in the upgrade path
	normFoundVersion, err := util.NormalizeVersion(bk.Status.CurrentVersion)
	if err != nil {
		// It should never happen
		return fmt.Errorf("found version is not in valid format, something bad happens: %v", err)
	}

	log.Printf("validateBookkeeperVersion:: normFoundVersion %s", normFoundVersion)
	upgradeString, ok := supportedVersions[normFoundVersion]
	if !ok {
		// It should never happen
		return fmt.Errorf("failed to find current cluster version in the supported versions")
	}
	upgradeList := strings.Split(upgradeString, ",")
	if !util.ContainsVersion(upgradeList, normRequestVersion) {
		return fmt.Errorf("unsupported upgrade from version %s to %s", bk.Status.CurrentVersion, requestVersion)
	}
	log.Print("validateBookkeeperVersion:: No error found...returning...")
	return nil
}

func (bk *BookkeeperCluster) LabelsForBookie() map[string]string {
	labels := bk.LabelsForBookkeeperCluster()
	if bk.Spec.Labels != nil {
		for k, v := range bk.Spec.Labels {
			labels[k] = v
		}
	}
	labels["component"] = "bookie"
	return labels
}

func (bookkeeperCluster *BookkeeperCluster) LabelsForBookkeeperCluster() map[string]string {
	return map[string]string{
		"app":                "bookkeeper-cluster",
		"bookkeeper_cluster": bookkeeperCluster.Name,
	}
}
func (bk *BookkeeperCluster) GetClusterExpectedSize() (size int) {
	return int(bk.Spec.Replicas)
}

func (bk *BookkeeperCluster) BookkeeperImage() (image string) {
	return fmt.Sprintf("%s:%s", bk.Spec.Image.Repository, bk.Spec.Version)
}

func (bk *BookkeeperCluster) BookkeeperTargetImage() (string, error) {
	if bk.Status.TargetVersion == "" {
		return "", fmt.Errorf("target version is not set")
	}
	return fmt.Sprintf("%s:%s", bk.Spec.Image.Repository, bk.Status.TargetVersion), nil
}

// Wait for pods in cluster to be terminated
func (bk *BookkeeperCluster) WaitForClusterToTerminate(kubeClient client.Client) (err error) {
	listOptions := &client.ListOptions{
		Namespace:     bk.Namespace,
		LabelSelector: labels.SelectorFromSet(bk.LabelsForBookkeeperCluster()),
	}
	err = wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		podList := &corev1.PodList{}
		err = kubeClient.List(context.TODO(), podList, listOptions)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}

		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	return err
}

func (bk *BookkeeperCluster) validateConfigMap() error {
	configmap := &corev1.ConfigMap{}
	err := Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: util.ConfigMapNameForBookie(bk.Name), Namespace: bk.Namespace}, configmap)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		} else {
			return fmt.Errorf("failed to get configmap (%s): %v", configmap.Name, err)
		}
	}
	if val, ok := bk.Spec.Options["journalDirectories"]; ok {
		eq := configmap.Data["BK_journalDirectories"] == val
		if !eq {
			return fmt.Errorf("path of journal directories should not be changed ")
		}
	}
	if val, ok := bk.Spec.Options["ledgerDirectories"]; ok {
		eq := configmap.Data["BK_ledgerDirectories"] == val
		if !eq {
			return fmt.Errorf("path of ledger directories should not be changed ")
		}
	}
	if val, ok := bk.Spec.Options["indexDirectories"]; ok {
		eq := configmap.Data["BK_indexDirectories"] == val
		if !eq {
			return fmt.Errorf("path of index directories should not be changed ")
		}
	}
	log.Print("validateConfigMap:: No error found...returning...")
	return nil
}
func (bk *BookkeeperCluster) NewEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	generateName := name + "-"
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    bk.Namespace,
			Labels:       bk.LabelsForBookkeeperCluster(),
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion:      "bookkeeper.pravega.io/v1alpha1",
			Kind:            "BookkeeperCluster",
			Name:            bk.GetName(),
			Namespace:       bk.GetNamespace(),
			ResourceVersion: bk.GetResourceVersion(),
			UID:             bk.GetUID(),
		},
		Reason:              reason,
		Message:             message,
		FirstTimestamp:      now,
		LastTimestamp:       now,
		Type:                eventType,
		ReportingController: operatorName,
		ReportingInstance:   os.Getenv("POD_NAME"),
	}
	return &event
}

func (bk *BookkeeperCluster) NewApplicationEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	generateName := name + "-"
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    bk.Namespace,
			Labels:       bk.LabelsForBookkeeperCluster(),
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: "app.k8s.io/v1beta1",
			Kind:       "Application",
			Name:       "bookkeeper-cluster",
			Namespace:  bk.GetNamespace(),
		},
		Reason:              reason,
		Message:             message,
		FirstTimestamp:      now,
		LastTimestamp:       now,
		Type:                eventType,
		ReportingController: operatorName,
		ReportingInstance:   os.Getenv("POD_NAME"),
	}
	return &event
}
