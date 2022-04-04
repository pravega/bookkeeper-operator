/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pravega/bookkeeper-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var bookkeeperclusterlog = logf.Log.WithName("bookkeepercluster-resource")

func (r *BookkeeperCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-bookkeeper-pravega-io-v1alpha1-bookkeepercluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=bookkeeper.pravega.io,resources=bookkeeperclusters,verbs=create;update,versions=v1alpha1,name=mbookkeepercluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &BookkeeperCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (bk *BookkeeperCluster) Default() {
	bookkeeperclusterlog.Info("default", "name", bk.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-bookkeeper-pravega-io-v1alpha1-bookkeepercluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=bookkeeper.pravega.io,resources=bookkeeperclusters,verbs=create;update,versions=v1alpha1,name=vbookkeepercluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &BookkeeperCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (bk *BookkeeperCluster) ValidateCreate() error {
	bookkeeperclusterlog.Info("validate create", "name", bk.Name)

	err := bk.ValidateBookkeeperVersion()
	if err != nil {
		return err
	}
	err = bk.ValidateAbsolutePath([]string{"journalDirectories", "ledgerDirectories", "indexDirectories"})
	if err != nil {
		return err
	}
	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (bk *BookkeeperCluster) ValidateUpdate(old runtime.Object) error {
	bookkeeperclusterlog.Info("validate update", "name", bk.Name)

	err := bk.ValidateBookkeeperVersion()
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
	bookkeeperclusterlog.Info("validate delete", "name", bk.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (bk *BookkeeperCluster) ValidateBookkeeperVersion() error {

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
	if match, _ := util.CompareVersions(normRequestVersion, normFoundVersion, "<"); match {
		return fmt.Errorf("downgrading the cluster from version %s to %s is not supported", bk.Status.CurrentVersion, requestVersion)
	}
	log.Printf("validateBookkeeperVersion:: normFoundVersion %s", normFoundVersion)

	log.Print("validateBookkeeperVersion:: No error found...returning...")
	return nil
}

func (bk *BookkeeperCluster) ValidateAbsolutePath(dirs []string) error {
	for _, dir := range dirs {
		if val, ok := bk.Spec.Options[dir]; ok {
			paths := strings.Split(val, ",")
			for _, path := range paths {
				if !strings.HasPrefix(path, "/") {
					return fmt.Errorf("path (%s) of %s should start with /", path, dir)
				}
			}
		}
	}
	return nil
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
	if val, ok := bk.Spec.Options["useHostNameAsBookieID"]; ok {
		eq := configmap.Data["BK_useHostNameAsBookieID"] == val
		if !eq {
			return fmt.Errorf("value of useHostNameAsBookieID should not be changed")
		}
	}
	if val, ok := bk.Spec.Options["journalDirectories"]; ok {
		eq := configmap.Data["BK_journalDirectories"] == val
		if !eq {
			return fmt.Errorf("path of journal directories should not be changed")
		}
	}
	if val, ok := bk.Spec.Options["ledgerDirectories"]; ok {
		eq := configmap.Data["BK_ledgerDirectories"] == val
		if !eq {
			return fmt.Errorf("path of ledger directories should not be changed")
		}
	}
	if val, ok := bk.Spec.Options["indexDirectories"]; ok {
		eq := configmap.Data["BK_indexDirectories"] == val
		if !eq {
			return fmt.Errorf("path of index directories should not be changed")
		}
	}
	log.Print("validateConfigMap:: No error found...returning...")
	return nil
}
