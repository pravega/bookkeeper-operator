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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/bookkeeper-operator/api/v1alpha1"
)

var _ = Describe("BookkeeperCluster Status", func() {

	var bk v1alpha1.BookkeeperCluster

	BeforeEach(func() {
		bk = v1alpha1.BookkeeperCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
	})

	Context("checking for default values", func() {
		BeforeEach(func() {
			bk.Status.VersionHistory = nil
			bk.Status.CurrentVersion = "0.5.0"
			bk.Status.Init()
		})
		It("should contains pods ready condition and it is false status", func() {
			_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionPodsReady)
			Ω(condition.Status).To(Equal(corev1.ConditionFalse))
		})
		It("should contains upgrade ready condition and it is false status", func() {
			_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionUpgrading)
			Ω(condition.Status).To(Equal(corev1.ConditionFalse))
		})
		It("should contains pods ready condition and it is false status", func() {
			_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionError)
			Ω(condition.Status).To(Equal(corev1.ConditionFalse))
		})
	})
	Context("checking for version history", func() {
		BeforeEach(func() {
			bk.Status.CurrentVersion = "0.7.0"
			bk.Status.Init()
			bk.Status.AddToVersionHistory("0.6.0")
		})
		It("version should get update correctly", func() {
			lastVersion := bk.Status.GetLastVersion()
			Ω(lastVersion).To(Equal("0.6.0"))
		})
	})
	Context("manually set pods ready condition to be true", func() {
		BeforeEach(func() {
			condition := v1alpha1.ClusterCondition{
				Type:               v1alpha1.ClusterConditionPodsReady,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			bk.Status.Conditions = append(bk.Status.Conditions, condition)
		})

		It("should contains pods ready condition and it is true status", func() {
			_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionPodsReady)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})

	Context("manually set pods upgrade condition to be true", func() {
		BeforeEach(func() {
			condition := v1alpha1.ClusterCondition{
				Type:               v1alpha1.ClusterConditionUpgrading,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			bk.Status.Conditions = append(bk.Status.Conditions, condition)
		})

		It("should contains pods upgrade condition and it is true status", func() {
			_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionUpgrading)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})
	Context("manually set pods Error condition to be true", func() {
		BeforeEach(func() {
			condition := v1alpha1.ClusterCondition{
				Type:               v1alpha1.ClusterConditionError,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			bk.Status.Conditions = append(bk.Status.Conditions, condition)
		})

		It("should contains pods error condition and it is true status", func() {
			_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionError)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})
	Context("set conditions", func() {
		Context("set pods ready condition to be true", func() {
			BeforeEach(func() {
				bk.Status.SetPodsReadyConditionFalse()
				bk.Status.SetPodsReadyConditionTrue()
			})
			It("should have pods ready condition with true status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionPodsReady)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))
			})
			It("should have pods ready condition with true status using function", func() {
				Ω(bk.Status.IsClusterInReadyState()).To(Equal(true))
			})
		})

		Context("set pod ready condition to be false", func() {
			BeforeEach(func() {
				bk.Status.SetPodsReadyConditionTrue()
				bk.Status.SetPodsReadyConditionFalse()
			})

			It("should have ready condition with false status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionPodsReady)
				Ω(condition.Status).To(Equal(corev1.ConditionFalse))
			})
			It("should have ready condition with false status using function", func() {
				Ω(bk.Status.IsClusterInReadyState()).To(Equal(false))
			})
			It("should have updated timestamps", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionPodsReady)
				// check the timestamps
				Ω(condition.LastUpdateTime).NotTo(Equal(""))
				Ω(condition.LastTransitionTime).NotTo(Equal(""))
			})
		})
		Context("set pod upgrade condition to be true", func() {
			BeforeEach(func() {
				bk.Status.SetUpgradingConditionFalse()
				bk.Status.SetUpgradingConditionTrue(" ", " ")
				bk.Status.UpdateProgress("UpdatingBookkeeperReason", "0")
			})
			It("should have pod upgrade condition with true status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionUpgrading)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))
				Ω(condition.Message).To(Equal("0"))
				Ω(condition.Reason).To(Equal("UpdatingBookkeeperReason"))
			})
			It("should have pod upgrade condition with true status using function", func() {
				Ω(bk.Status.IsClusterInUpgradingState()).To(Equal(true))
			})
			It("Checking GetlastCondition function and It should return UpgradeCondition as cluster in Upgrading state", func() {
				condition := bk.Status.GetLastCondition()
				Ω(string(condition.Type)).To(Equal(v1alpha1.ClusterConditionUpgrading))
			})
			It("Checking ClusterInUpgradeFailedOrRollbackState should return false ", func() {
				Ω(bk.Status.IsClusterInUpgradeFailedOrRollbackState()).To(Equal(false))
			})
			It("Checking ClusterInRollbackFailedState should return false", func() {
				Ω(bk.Status.IsClusterInRollbackFailedState()).To(Equal(false))
			})
		})
		Context("set pod upgrade condition to be false", func() {
			BeforeEach(func() {
				bk.Status.SetUpgradingConditionTrue(" ", " ")
				bk.Status.SetUpgradingConditionFalse()
			})

			It("should have upgrade condition with false status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionUpgrading)
				Ω(condition.Status).To(Equal(corev1.ConditionFalse))
			})

			It("should have upgrade condition with false status using function", func() {
				Ω(bk.Status.IsClusterInUpgradingState()).To(Equal(false))
			})

			It("Checking GetlastCondition function and It should return nil as not in Upgrading state", func() {
				condition := bk.Status.GetLastCondition()
				Ω(condition).To(BeNil())
			})

			It("should have updated timestamps", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionUpgrading)
				//check the timestamps
				Ω(condition.LastUpdateTime).NotTo(Equal(""))
				Ω(condition.LastTransitionTime).NotTo(Equal(""))
			})
		})
		Context("set pods Error condition  upgrade failed to be true", func() {
			BeforeEach(func() {
				bk.Status.SetErrorConditionFalse()
				bk.Status.SetErrorConditionTrue("UpgradeFailed", " ")
			})
			It("should have pods Error condition with true status using function", func() {
				Ω(bk.Status.IsClusterInUpgradeFailedState()).To(Equal(true))
			})
			It("should have pods Error condition with true status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionError)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))
			})
			It("Checking ClusterInUpgradeFailedOrRollbackState and It should return true", func() {
				Ω(bk.Status.IsClusterInUpgradeFailedOrRollbackState()).To(Equal(true))
			})
		})

		Context("set pods Error condition  rollback failed to be true", func() {
			BeforeEach(func() {
				bk.Status.SetErrorConditionFalse()
				bk.Status.SetErrorConditionTrue("RollbackFailed", " ")
			})
			It("should return rollback failed state to true using function", func() {
				Ω(bk.Status.IsClusterInRollbackFailedState()).To(Equal(true))
			})
			It("should return rollback failed state to false using function", func() {
				bk.Status.SetErrorConditionTrue("some err", "")
				Ω(bk.Status.IsClusterInRollbackFailedState()).To(Equal(false))
			})

			It("should have pods Error condition with true status using function", func() {
				Ω(bk.Status.IsClusterInErrorState()).To(Equal(true))
			})
			It("should have pods Error condition with true status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionError)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))

			})
		})
		Context("set pod Error condition to be false", func() {
			BeforeEach(func() {
				bk.Status.SetErrorConditionTrue("UpgradeFailed", " ")
				bk.Status.SetErrorConditionFalse()
			})

			It("should have Error condition with false status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionError)
				Ω(condition.Status).To(Equal(corev1.ConditionFalse))
			})

			It("should have Error condition with false status using function", func() {
				Ω(bk.Status.IsClusterInUpgradeFailedState()).To(Equal(false))
			})

			It("cluster in error state should return false", func() {
				Ω(bk.Status.IsClusterInErrorState()).To(Equal(false))
			})

			It("should have updated timestamps", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionError)
				//check the timestamps
				Ω(condition.LastUpdateTime).NotTo(Equal(""))
				Ω(condition.LastTransitionTime).NotTo(Equal(""))
			})
		})
		Context("set pods rollback condition to be true", func() {
			BeforeEach(func() {
				bk.Status.SetRollbackConditionFalse()
				bk.Status.SetRollbackConditionTrue(" ", " ")
				bk.Status.UpdateProgress("UpgradeErrorReason", "")
			})
			It("should have pods rollback condition with true status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionRollback)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))
			})
			It("should have pods rollback condition with true status using function", func() {
				Ω(bk.Status.IsClusterInRollbackState()).To(Equal(true))
			})
			It("Checking GetlastCondition function and It should return RollbackCondition as cluster in Rollback state", func() {
				condition := bk.Status.GetLastCondition()
				Ω(string(condition.Type)).To(Equal(v1alpha1.ClusterConditionRollback))
			})
			It("Checking ClusterInUpgradeFailedOrRollbackState and It should return true", func() {
				Ω(bk.Status.IsClusterInUpgradeFailedOrRollbackState()).To(Equal(true))
			})
		})
		Context("set pods rollback condition to be false", func() {
			BeforeEach(func() {
				bk.Status.SetRollbackConditionTrue(" ", " ")
				bk.Status.SetRollbackConditionFalse()
			})
			It("should have pods rollback condition with false status", func() {
				_, condition := bk.Status.GetClusterCondition(v1alpha1.ClusterConditionRollback)
				Ω(condition.Status).To(Equal(corev1.ConditionFalse))
			})
			It("should have pods rollback condition with false status using function", func() {
				Ω(bk.Status.IsClusterInRollbackState()).To(Equal(false))
			})
			It("Checking GetlastCondition function and It should return nil n as cluster not in Rollback state", func() {
				condition := bk.Status.GetLastCondition()
				Ω(condition).To(BeNil())
			})
		})
	})
})
