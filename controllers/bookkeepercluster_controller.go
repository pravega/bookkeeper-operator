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

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	//	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bookkeeperv1alpha1 "github.com/pravega/bookkeeper-operator/api/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/controller/config"
	"github.com/pravega/bookkeeper-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

var _ reconcile.Reconciler = &BookkeeperClusterReconciler{}

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

// BookkeeperClusterReconciler reconciles a BookkeeperCluster object
type BookkeeperClusterReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=bookkeeper.pravega.io,resources=bookkeeperclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bookkeeper.pravega.io,resources=bookkeeperclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bookkeeper.pravega.io,resources=bookkeeperclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BookkeeperCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile

func (r *BookkeeperClusterReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	//	_ = log.FromContext(ctx)

	log.Printf("Reconciling BookkeeperCluster %s/%s\n", request.Namespace, request.Name)

	// Fetch the BookkeeperCluster instance
	bookkeeperCluster := &bookkeeperv1alpha1.BookkeeperCluster{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, bookkeeperCluster)
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
		if err = r.Client.Update(context.TODO(), bookkeeperCluster); err != nil {
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
func (r *BookkeeperClusterReconciler) run(p *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
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

func (r *BookkeeperClusterReconciler) deployCluster(p *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	err = r.deployBookie(p)
	if err != nil {
		log.Printf("failed to deploy bookie: %v", err)
		return err
	}
	return nil
}

func (r *BookkeeperClusterReconciler) deployBookie(p *bookkeeperv1alpha1.BookkeeperCluster) (err error) {

	statefulSet := MakeBookieStatefulSet(p)
	controllerutil.SetControllerReference(p, statefulSet, r.Scheme)
	for i := range statefulSet.Spec.VolumeClaimTemplates {
		controllerutil.SetControllerReference(p, &statefulSet.Spec.VolumeClaimTemplates[i], r.Scheme)
		if *p.Spec.BlockOwnerDeletion == false {
			refs := statefulSet.Spec.VolumeClaimTemplates[i].OwnerReferences
			for i := range refs {
				refs[i].BlockOwnerDeletion = p.Spec.BlockOwnerDeletion
			}
		}
	}
	err = r.Client.Create(context.TODO(), statefulSet)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		} else {
			sts := &appsv1.StatefulSet{}
			name := util.StatefulSetNameForBookie(p.Name)
			err := r.Client.Get(context.TODO(),
				types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
			if err != nil {
				return err
			}
			if !r.checkVersionUpgradeTriggered(p) && !r.isRollbackTriggered(p) {
				originalsts := sts.DeepCopy()
				sts.Spec.Template = statefulSet.Spec.Template
				err = r.Client.Update(context.TODO(), sts)
				if err != nil {
					return fmt.Errorf("failed to update stateful set: %v", err)
				}
				if !reflect.DeepEqual(originalsts.Spec.Template, sts.Spec.Template) {
					err = r.restartStsPod(p)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (r *BookkeeperClusterReconciler) syncClusterSize(p *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	err = r.syncBookieSize(p)
	if err != nil {
		return err
	}
	return nil
}

func (r *BookkeeperClusterReconciler) syncBookieSize(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	sts := &appsv1.StatefulSet{}
	name := util.StatefulSetNameForBookie(bk.Name)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: bk.Namespace}, sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != bk.Spec.Replicas {
		sts.Spec.Replicas = &(bk.Spec.Replicas)
		err = r.Client.Update(context.TODO(), sts)
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

func (r *BookkeeperClusterReconciler) reconcileFinalizers(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	currentBookkeeperCluster := &bookkeeperv1alpha1.BookkeeperCluster{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bk.Name, Namespace: bk.Namespace}, currentBookkeeperCluster)
	if err != nil {
		return fmt.Errorf("failed to get bookkeeper cluster (%s): %v", bk.Name, err)
	}
	bk.ObjectMeta.ResourceVersion = currentBookkeeperCluster.ObjectMeta.ResourceVersion
	if bk.DeletionTimestamp.IsZero() && !config.DisableFinalizer {
		// checks whether the slice of finalizers contains a string with the given prefix
		// NOTE: we need to ensure that no two finalizer names have the same prefix
		if !util.ContainsStringWithPrefix(bk.ObjectMeta.Finalizers, util.ZkFinalizer) {
			finalizer := util.ZkFinalizer
			configMap := &corev1.ConfigMap{}
			if strings.TrimSpace(bk.Spec.EnvVars) != "" {
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: strings.TrimSpace(bk.Spec.EnvVars), Namespace: bk.Namespace}, configMap)
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
			if err = r.Client.Update(context.TODO(), bk); err != nil {
				return fmt.Errorf("failed to add the finalizer (%s): %v", bk.Name, err)
			}
		}
	} else {
		// checks whether the slice of finalizers contains a string with the given prefix
		if util.ContainsStringWithPrefix(bk.ObjectMeta.Finalizers, util.ZkFinalizer) {
			finalizer, pravegaClusterName := getFinalizerAndClusterName(bk.ObjectMeta.Finalizers)
			bk.ObjectMeta.Finalizers = util.RemoveString(bk.ObjectMeta.Finalizers, finalizer)
			if err = r.Client.Update(context.TODO(), bk); err != nil {
				return fmt.Errorf("failed to update Bookkeeper object (%s): %v", bk.Name, err)
			}
			if err = r.cleanUpZookeeperMeta(bk, pravegaClusterName); err != nil {
				// emit an event for zk metadata cleanup failure
				message := fmt.Sprintf("failed to cleanup %s metadata from zookeeper (znode path: /pravega/%s): %v", bk.Name, pravegaClusterName, err)
				event := bk.NewApplicationEvent("ZKMETA_CLEANUP_ERROR", "ZK Metadata Cleanup Failed", message, "Error")
				pubErr := r.Client.Create(context.TODO(), event)
				if pubErr != nil {
					log.Printf("Error publishing zk metadata cleanup failure event to k8s. %v", pubErr)
				}
				return fmt.Errorf(message)
			}
		}
	}
	return nil
}

func (r *BookkeeperClusterReconciler) cleanUpZookeeperMeta(bk *bookkeeperv1alpha1.BookkeeperCluster, pravegaClusterName string) (err error) {
	if err = bk.WaitForClusterToTerminate(r.Client); err != nil {
		return fmt.Errorf("failed to wait for cluster pods termination (%s): %v", bk.Name, err)
	}

	if err = util.DeleteAllZnodes(bk.Spec.ZookeeperUri, bk.Namespace, pravegaClusterName); err != nil {
		return fmt.Errorf("failed to delete zookeeper znodes for (%s): %v", bk.Name, err)
	}
	return nil
}

func (r *BookkeeperClusterReconciler) syncStatefulSetPvc(sts *appsv1.StatefulSet) error {
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
	err = r.Client.List(context.TODO(), pvcList, pvclistOps)
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

			err = r.Client.Delete(context.TODO(), pvcDelete)
			if err != nil {
				return fmt.Errorf("failed to delete pvc: %v", err)
			}
		}
	}
	return nil
}
func (r *BookkeeperClusterReconciler) reconcileConfigMap(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {

	currentConfigMap := &corev1.ConfigMap{}
	configMap := MakeBookieConfigMap(bk)
	controllerutil.SetControllerReference(bk, configMap, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: util.ConfigMapNameForBookie(bk.Name), Namespace: bk.Namespace}, currentConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), configMap)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	} else {
		currentConfigMap := &corev1.ConfigMap{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: util.ConfigMapNameForBookie(bk.Name), Namespace: bk.Namespace}, currentConfigMap)
		eq := util.CompareConfigMap(currentConfigMap, configMap)
		if !eq {
			err := r.Client.Update(context.TODO(), configMap)
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

func (r *BookkeeperClusterReconciler) checkVersionUpgradeTriggered(bk *bookkeeperv1alpha1.BookkeeperCluster) bool {
	currentBookkeeperCluster := &bookkeeperv1alpha1.BookkeeperCluster{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: bk.Name, Namespace: bk.Namespace}, currentBookkeeperCluster)
	if err == nil && currentBookkeeperCluster.Status.CurrentVersion != bk.Spec.Version {
		return true
	}
	return false
}
func (r *BookkeeperClusterReconciler) reconcilePdb(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {

	pdb := MakeBookiePodDisruptionBudget(bk)
	controllerutil.SetControllerReference(bk, pdb, r.Scheme)
	err = r.Client.Create(context.TODO(), pdb)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	currentPdb := &policyv1beta1.PodDisruptionBudget{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: util.PdbNameForBookie(bk.Name), Namespace: bk.Namespace}, currentPdb)
	if err != nil {
		return err
	}
	return r.updatePdb(currentPdb, pdb)
}

func (r *BookkeeperClusterReconciler) updatePdb(currentPdb *policyv1beta1.PodDisruptionBudget, newPdb *policyv1beta1.PodDisruptionBudget) (err error) {

	if !reflect.DeepEqual(currentPdb.Spec.MaxUnavailable, newPdb.Spec.MaxUnavailable) {
		currentPdb.Spec.MaxUnavailable = newPdb.Spec.MaxUnavailable
		err = r.Client.Update(context.TODO(), currentPdb)
		if err != nil {
			return fmt.Errorf("failed to update pdb (%s): %v", currentPdb.Name, err)
		}
	}
	return nil
}

func (r *BookkeeperClusterReconciler) reconcileService(bk *bookkeeperv1alpha1.BookkeeperCluster) error {
	headlessService := MakeBookieHeadlessService(bk)
	controllerutil.SetControllerReference(bk, headlessService, r.Scheme)
	err := r.Client.Create(context.TODO(), headlessService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	currentPdb := &policyv1beta1.PodDisruptionBudget{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: util.PdbNameForBookie(bk.Name), Namespace: bk.Namespace}, currentPdb)

	return nil
}

func (r *BookkeeperClusterReconciler) restartStsPod(bk *bookkeeperv1alpha1.BookkeeperCluster) error {

	currentSts := &appsv1.StatefulSet{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: util.StatefulSetNameForBookie(bk.Name), Namespace: bk.Namespace}, currentSts)
	if err != nil {
		return err
	}
	labels := bk.LabelsForBookkeeperCluster()
	labels["component"] = "bookie"
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}
	podList := &corev1.PodList{}
	podlistOps := &client.ListOptions{
		Namespace:     currentSts.Namespace,
		LabelSelector: selector,
	}
	err = r.Client.List(context.TODO(), podList, podlistOps)
	if err != nil {
		return err
	}
	sort.SliceStable(podList.Items, func(i int, j int) bool {
		return podList.Items[i].Name < podList.Items[j].Name
	})
	for _, podItem := range podList.Items {
		err := r.Client.Delete(context.TODO(), &podItem)
		if err != nil {
			return err
		} else {
			start := time.Now()
			pod := &corev1.Pod{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to delete Bookkeeper pod (%s) for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
			start = time.Now()
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for !util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to get Bookkeeper pod (%s) as ready for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
		}
	}
	return nil
}

func (r *BookkeeperClusterReconciler) syncStatefulSetExternalServices(sts *appsv1.StatefulSet) error {
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
	err = r.Client.List(context.TODO(), serviceList, servicelistOps)
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

			err = r.Client.Delete(context.TODO(), svcDelete)
			if err != nil {
				return fmt.Errorf("failed to delete svc: %v", err)
			}
		}
	}
	return nil
}

func (r *BookkeeperClusterReconciler) reconcileClusterStatus(bk *bookkeeperv1alpha1.BookkeeperCluster) error {

	bk.Status.Init()

	expectedSize := bk.GetClusterExpectedSize()
	listOps := &client.ListOptions{
		Namespace:     bk.Namespace,
		LabelSelector: labels.SelectorFromSet(bk.LabelsForBookkeeperCluster()),
	}
	podList := &corev1.PodList{}
	err := r.Client.List(context.TODO(), podList, listOps)
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

	err = r.Client.Status().Update(context.TODO(), bk)
	if err != nil {
		return fmt.Errorf("failed to update cluster status: %v", err)
	}
	return nil
}

func (r *BookkeeperClusterReconciler) rollbackFailedUpgrade(bk *bookkeeperv1alpha1.BookkeeperCluster) error {
	if r.isRollbackTriggered(bk) {
		// start rollback to previous version
		previousVersion := bk.Status.GetLastVersion()
		log.Printf("Rolling back to last cluster version  %v", previousVersion)
		//Rollback cluster to previous version
		return r.rollbackClusterVersion(bk, previousVersion)
	}
	return nil
}

func (r *BookkeeperClusterReconciler) isRollbackTriggered(bk *bookkeeperv1alpha1.BookkeeperCluster) bool {
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

// SetupWithManager sets up the controller with the Manager.
func (r *BookkeeperClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bookkeeperv1alpha1.BookkeeperCluster{}).
		Complete(r)
}
