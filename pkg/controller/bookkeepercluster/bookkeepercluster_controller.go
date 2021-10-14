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
	"context"
	"fmt"
	"github.com/pravega/bookkeeper-operator/pkg/controller/config"
	"reflect"
	"sort"
	"strings"
	"time"

	bookkeeperv1alpha1 "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	log "github.com/sirupsen/logrus"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

// Add creates a new bookkeeperCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBookkeeperCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("bookkeeper-cluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource bookkeeperCluster
	err = c.Watch(&source.Kind{Type: &bookkeeperv1alpha1.BookkeeperCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileBookkeeperCluster{}

// ReconcileBookkeeperCluster reconciles a BookkeeperCluster object
type ReconcileBookkeeperCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BookkeeperCluster object and makes changes based on the state read
// and what is in the BookkeeperCluster.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBookkeeperCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling BookkeeperCluster %s/%s\n", request.Namespace, request.Name)

	// Fetch the BookkeeperCluster instance
	bookkeeperCluster := &bookkeeperv1alpha1.BookkeeperCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, bookkeeperCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Printf("BookkeeperCluster %s/%s not found. Ignoring since object must be deleted\n", request.Namespace, request.Name)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Printf("failed to get BookkeeperCluster: %v", err)
		return reconcile.Result{}, err
	}

	// Set default configuration for unspecified values
	changed := bookkeeperCluster.WithDefaults()
	if changed {
		log.Printf("Setting default settings for bookkeeper-cluster: %s", request.Name)
		if err = r.client.Update(context.TODO(), bookkeeperCluster); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	err = r.run(bookkeeperCluster)
	if err != nil {
		log.Printf("failed to reconcile bookkeeper cluster (%s): %v", bookkeeperCluster.Name, err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}

func (r *ReconcileBookkeeperCluster) run(p *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	// Clean up zookeeper metadata
	err = r.reconcileFinalizers(p)
	if err != nil {
		return fmt.Errorf("failed to clean up zookeeper: %v", err)
	}

	err = r.reconcileConfigMap(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile configMap %v", err)
	}

	err = r.reconcilePdb(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile pdb %v", err)
	}

	err = r.reconcileService(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile service %v", err)
	}

	err = r.deployCluster(p)
	if err != nil {
		return fmt.Errorf("failed to deploy cluster: %v", err)
	}

	err = r.syncClusterSize(p)
	if err != nil {
		return fmt.Errorf("failed to sync cluster size: %v", err)
	}

	// Upgrade
	err = r.syncClusterVersion(p)
	if err != nil {
		return fmt.Errorf("failed to sync cluster version: %v", err)
	}

	// Rollback
	err = r.rollbackFailedUpgrade(p)
	if err != nil {
		return fmt.Errorf("Rollback attempt failed: %v", err)
	}

	err = r.reconcileClusterStatus(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile cluster status: %v", err)
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) deployCluster(p *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	err = r.deployBookie(p)
	if err != nil {
		log.Printf("failed to deploy bookie: %v", err)
		return err
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) deployBookie(p *bookkeeperv1alpha1.BookkeeperCluster) (err error) {

	statefulSet := MakeBookieStatefulSet(p)
	controllerutil.SetControllerReference(p, statefulSet, r.scheme)
	for i := range statefulSet.Spec.VolumeClaimTemplates {
		controllerutil.SetControllerReference(p, &statefulSet.Spec.VolumeClaimTemplates[i], r.scheme)
		if *p.Spec.BlockOwnerDeletion == false {
			refs := statefulSet.Spec.VolumeClaimTemplates[i].OwnerReferences
			for i := range refs {
				refs[i].BlockOwnerDeletion = p.Spec.BlockOwnerDeletion
			}
		}
	}
	err = r.client.Create(context.TODO(), statefulSet)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *ReconcileBookkeeperCluster) syncClusterSize(p *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	err = r.syncBookieSize(p)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) syncBookieSize(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	sts := &appsv1.StatefulSet{}
	name := util.StatefulSetNameForBookie(bk.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: bk.Namespace}, sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != bk.Spec.Replicas {
		sts.Spec.Replicas = &(bk.Spec.Replicas)
		err = r.client.Update(context.TODO(), sts)
		if err != nil {
			return fmt.Errorf("failed to update size of stateful-set (%s): %v", sts.Name, err)
		}

		err = r.syncStatefulSetPvc(sts)
		if err != nil {
			return fmt.Errorf("failed to sync pvcs of stateful-set (%s): %v", sts.Name, err)
		}
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) reconcileFinalizers(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	if bk.DeletionTimestamp.IsZero() && !config.DisableFinalizer {
		// checks whether the slice of finalizers contains a string with the given prefix
		// NOTE: we need to ensure that no two finalizer names have the same prefix
		if !util.ContainsStringWithPrefix(bk.ObjectMeta.Finalizers, util.ZkFinalizer) {
			finalizer := util.ZkFinalizer
			configMap := &corev1.ConfigMap{}
			if strings.TrimSpace(bk.Spec.EnvVars) != "" {
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: strings.TrimSpace(bk.Spec.EnvVars), Namespace: bk.Namespace}, configMap)
				if err != nil {
					return fmt.Errorf("failed to get the configmap %s: %v", bk.Spec.EnvVars, err)
				}
				clusterName, ok := configMap.Data["PRAVEGA_CLUSTER_NAME"]
				if ok {
					// appending name of pravega cluster to the name of the finalizer
					// to handle zk metadata deletion
					finalizer = finalizer + "_" + clusterName
				}
			}
			bk.ObjectMeta.Finalizers = append(bk.ObjectMeta.Finalizers, finalizer)
			if err = r.client.Update(context.TODO(), bk); err != nil {
				return fmt.Errorf("failed to add the finalizer (%s): %v", bk.Name, err)
			}
		}
	} else {
		// checks whether the slice of finalizers contains a string with the given prefix
		if util.ContainsStringWithPrefix(bk.ObjectMeta.Finalizers, util.ZkFinalizer) {
			finalizer, pravegaClusterName := getFinalizerAndClusterName(bk.ObjectMeta.Finalizers)
			bk.ObjectMeta.Finalizers = util.RemoveString(bk.ObjectMeta.Finalizers, finalizer)
			if err = r.client.Update(context.TODO(), bk); err != nil {
				return fmt.Errorf("failed to update Bookkeeper object (%s): %v", bk.Name, err)
			}
			if err = r.cleanUpZookeeperMeta(bk, pravegaClusterName); err != nil {
				// emit an event for zk metadata cleanup failure
				message := fmt.Sprintf("failed to cleanup %s metadata from zookeeper (znode path: /pravega/%s): %v", bk.Name, pravegaClusterName, err)
				event := bk.NewApplicationEvent("ZKMETA_CLEANUP_ERROR", "ZK Metadata Cleanup Failed", message, "Error")
				pubErr := r.client.Create(context.TODO(), event)
				if pubErr != nil {
					log.Printf("Error publishing zk metadata cleanup failure event to k8s. %v", pubErr)
				}
				return fmt.Errorf(message)
			}
		}
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) cleanUpZookeeperMeta(bk *bookkeeperv1alpha1.BookkeeperCluster, pravegaClusterName string) (err error) {
	if err = bk.WaitForClusterToTerminate(r.client); err != nil {
		return fmt.Errorf("failed to wait for cluster pods termination (%s): %v", bk.Name, err)
	}

	if err = util.DeleteAllZnodes(bk.Spec.ZookeeperUri, bk.Namespace, pravegaClusterName); err != nil {
		return fmt.Errorf("failed to delete zookeeper znodes for (%s): %v", bk.Name, err)
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) syncStatefulSetPvc(sts *appsv1.StatefulSet) error {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: sts.Spec.Template.Labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}

	pvcList := &corev1.PersistentVolumeClaimList{}
	pvclistOps := &client.ListOptions{
		Namespace:     sts.Namespace,
		LabelSelector: selector,
	}
	err = r.client.List(context.TODO(), pvcList, pvclistOps)
	if err != nil {
		return err
	}

	for _, pvcItem := range pvcList.Items {
		if util.IsOrphan(pvcItem.Name, *sts.Spec.Replicas) {
			pvcDelete := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcItem.Name,
					Namespace: pvcItem.Namespace,
				},
			}

			err = r.client.Delete(context.TODO(), pvcDelete)
			if err != nil {
				return fmt.Errorf("failed to delete pvc: %v", err)
			}
		}
	}
	return nil
}
func (r *ReconcileBookkeeperCluster) reconcileConfigMap(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {

	currentConfigMap := &corev1.ConfigMap{}
	configMap := MakeBookieConfigMap(bk)
	controllerutil.SetControllerReference(bk, configMap, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: util.ConfigMapNameForBookie(bk.Name), Namespace: bk.Namespace}, currentConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), configMap)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	} else {
		currentConfigMap := &corev1.ConfigMap{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: util.ConfigMapNameForBookie(bk.Name), Namespace: bk.Namespace}, currentConfigMap)
		eq := util.CompareConfigMap(currentConfigMap, configMap)
		if !eq {
			err := r.client.Update(context.TODO(), configMap)
			if err != nil {
				return err
			}
			//restarting sts pods
			if !r.checkVersionUpgradeTriggered(bk) {
				err = r.restartStsPod(bk)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) checkVersionUpgradeTriggered(bk *bookkeeperv1alpha1.BookkeeperCluster) bool {
	currentBookkeeperCluster := &bookkeeperv1alpha1.BookkeeperCluster{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: bk.Name, Namespace: bk.Namespace}, currentBookkeeperCluster)
	if err == nil && currentBookkeeperCluster.Status.CurrentVersion != bk.Spec.Version {
		return true
	}
	return false
}
func (r *ReconcileBookkeeperCluster) reconcilePdb(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {

	pdb := MakeBookiePodDisruptionBudget(bk)
	controllerutil.SetControllerReference(bk, pdb, r.scheme)
	err = r.client.Create(context.TODO(), pdb)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	currentPdb := &policyv1beta1.PodDisruptionBudget{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: util.PdbNameForBookie(bk.Name), Namespace: bk.Namespace}, currentPdb)
	if err != nil {
		return err
	}
	return r.updatePdb(currentPdb, pdb)
}

func (r *ReconcileBookkeeperCluster) updatePdb(currentPdb *policyv1beta1.PodDisruptionBudget, newPdb *policyv1beta1.PodDisruptionBudget) (err error) {

	if !reflect.DeepEqual(currentPdb.Spec.MaxUnavailable, newPdb.Spec.MaxUnavailable) {
		currentPdb.Spec.MaxUnavailable = newPdb.Spec.MaxUnavailable
		err = r.client.Update(context.TODO(), currentPdb)
		if err != nil {
			return fmt.Errorf("failed to update pdb (%s): %v", currentPdb.Name, err)
		}
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) reconcileService(bk *bookkeeperv1alpha1.BookkeeperCluster) error {
	headlessService := MakeBookieHeadlessService(bk)
	controllerutil.SetControllerReference(bk, headlessService, r.scheme)
	err := r.client.Create(context.TODO(), headlessService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	currentPdb := &policyv1beta1.PodDisruptionBudget{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: util.PdbNameForBookie(bk.Name), Namespace: bk.Namespace}, currentPdb)

	return nil
}

func (r *ReconcileBookkeeperCluster) restartStsPod(bk *bookkeeperv1alpha1.BookkeeperCluster) error {

	currentSts := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: util.StatefulSetNameForBookie(bk.Name), Namespace: bk.Namespace}, currentSts)
	if err != nil {
		return err
	}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: currentSts.Spec.Template.Labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}
	podList := &corev1.PodList{}
	podlistOps := &client.ListOptions{
		Namespace:     currentSts.Namespace,
		LabelSelector: selector,
	}
	err = r.client.List(context.TODO(), podList, podlistOps)
	if err != nil {
		return err
	}
	sort.SliceStable(podList.Items, func(i int, j int) bool {
		return podList.Items[i].Name < podList.Items[j].Name
	})
	for _, podItem := range podList.Items {
		err := r.client.Delete(context.TODO(), &podItem)
		if err != nil {
			return err
		} else {
			start := time.Now()
			pod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to delete Bookkeeper pod (%s) for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
			start = time.Now()
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for !util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to get Bookkeeper pod (%s) as ready for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
		}
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) syncStatefulSetExternalServices(sts *appsv1.StatefulSet) error {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: sts.Spec.Template.Labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}

	serviceList := &corev1.ServiceList{}
	servicelistOps := &client.ListOptions{
		Namespace:     sts.Namespace,
		LabelSelector: selector,
	}
	err = r.client.List(context.TODO(), serviceList, servicelistOps)
	if err != nil {
		return err
	}
	for _, svcItem := range serviceList.Items {

		if util.IsOrphan(svcItem.Name, *sts.Spec.Replicas) {
			svcDelete := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      svcItem.Name,
					Namespace: svcItem.Namespace,
				},
			}

			err = r.client.Delete(context.TODO(), svcDelete)
			if err != nil {
				return fmt.Errorf("failed to delete svc: %v", err)
			}
		}
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) reconcileClusterStatus(bk *bookkeeperv1alpha1.BookkeeperCluster) error {

	bk.Status.Init()

	expectedSize := bk.GetClusterExpectedSize()
	listOps := &client.ListOptions{
		Namespace:     bk.Namespace,
		LabelSelector: labels.SelectorFromSet(bk.LabelsForBookkeeperCluster()),
	}
	podList := &corev1.PodList{}
	err := r.client.List(context.TODO(), podList, listOps)
	if err != nil {
		return err
	}

	var (
		readyMembers   []string
		unreadyMembers []string
	)

	for _, pod := range podList.Items {
		if util.IsPodReady(&pod) {
			readyMembers = append(readyMembers, pod.Name)
		} else {
			unreadyMembers = append(unreadyMembers, pod.Name)
		}
	}

	if len(readyMembers) == expectedSize {
		bk.Status.SetPodsReadyConditionTrue()
	} else {
		bk.Status.SetPodsReadyConditionFalse()
	}

	bk.Status.Replicas = int32(expectedSize)
	bk.Status.CurrentReplicas = int32(len(podList.Items))
	bk.Status.ReadyReplicas = int32(len(readyMembers))
	bk.Status.Members.Ready = readyMembers
	bk.Status.Members.Unready = unreadyMembers

	err = r.client.Status().Update(context.TODO(), bk)
	if err != nil {
		return fmt.Errorf("failed to update cluster status: %v", err)
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) rollbackFailedUpgrade(bk *bookkeeperv1alpha1.BookkeeperCluster) error {
	if r.isRollbackTriggered(bk) {
		// start rollback to previous version
		previousVersion := bk.Status.GetLastVersion()
		log.Printf("Rolling back to last cluster version  %v", previousVersion)
		//Rollback cluster to previous version
		return r.rollbackClusterVersion(bk, previousVersion)
	}
	return nil
}

func (r *ReconcileBookkeeperCluster) isRollbackTriggered(bk *bookkeeperv1alpha1.BookkeeperCluster) bool {
	if bk.Status.IsClusterInUpgradeFailedState() && bk.Spec.Version == bk.Status.GetLastVersion() {
		return true
	}
	return false
}

func getFinalizerAndClusterName(slice []string) (string, string) {
	// get the modified finalizer name from the slice of finalizers with the given prefix
	finalizer := util.GetStringWithPrefix(slice, util.ZkFinalizer)
	// extracting pravega cluster name from the modified finalizer name
	pravegaClusterName := strings.Replace(finalizer, util.ZkFinalizer, "", 1)
	if pravegaClusterName == "" {
		pravegaClusterName = "pravega-cluster"
	} else {
		pravegaClusterName = strings.Replace(pravegaClusterName, "_", "", 1)
	}
	return finalizer, pravegaClusterName
}
