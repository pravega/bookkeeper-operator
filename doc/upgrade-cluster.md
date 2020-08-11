# Bookkeeper cluster upgrade

This document shows how to upgrade a Pravega cluster managed by the operator to a desired version while preserving the cluster's state and data whenever possible.

## Overview

The activity diagram below shows the overall upgrade process started by an end-user and performed by the operator.

![pravega k8 upgrade 1](https://user-images.githubusercontent.com/3786750/51993601-7908b000-24af-11e9-8149-82fd1b036630.png)


## Prerequisites

Your Bookkeeper cluster should be in a healthy state. You can check your cluster health by listing it and checking that all members are ready.

```
$ kubectl get bk
NAME        VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
bookkeeper  0.4.0        7                 7            11m
```

## Valid Upgrade Paths

To understand the valid upgrade paths for a pravega cluster, refer to the [version map](https://github.com/pravega/bookkeeper-operator/blob/master/deploy/version_map.yaml). The key indicates the base version of the cluster, and the value against each key indicates the list of valid versions this base version can be upgraded to.

## Trigger an upgrade

### Upgrading via Helm

The upgrade can be triggered via helm using the following command
```
$ helm upgrade <bookkeeper cluster release name> <location of modified charts> --timeout 600s
```

### Upgrading manually

To initiate the upgrade process manually, a user has to update the `spec.version` field on the `BookkeeperCluster` custom resource. This can be done in three different ways using the `kubectl` command.
1. `kubectl edit BookkeeperCluster <name>`, modify the `version` value in the YAML resource, save, and exit.
2. If you have the custom resource defined in a local YAML file, e.g. `bookkeeper.yaml`, you can modify the `version` value, and reapply the resource with `kubectl apply -f bookkeeper.yaml`.
3. `kubectl patch BookkeeperCluster <name> --type='json' -p='[{"op": "replace", "path": "/spec/version", "value": "X.Y.Z"}]'`.
After the `version` field is updated, the operator will detect the version change and it will trigger the upgrade process.

## Upgrade process

Once an upgrade request has been received, the operator will apply the rolling upgrade to the Bookkeeper STS.

The upgrade workflow is as follows:

- The operator will change the `Upgrade` condition to `True` to indicate that the cluster resource has an upgrade in progress.
  - If any of the component pods has errors, the upgrade process will stop (`Upgrade` condition to `False`) and operator will set the `Error` condition to `True` and indicate the reason.
- When all pods are upgraded, the `Upgrade` condition will be set to `False` and `status.currentVersion` will be updated to the desired version.


### BookKeeper upgrade

BookKeeper cluster is deployed as a [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) due to its requirements on:

- Persistent storage: each bookie has three persistent volume for ledgers, journals, and indices. If a pod is migrated or recreated (e.g. when it's upgraded), the data in those volumes will remain untouched.
- Stable network names: the `StatefulSet` provides pods with a predictable name and a [Headless service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) creates DNS records for pods to be reachable by clients. If a pod is recreated or migrated to a different node, clients will continue to be able to reach the pod despite changing its IP address.

Statefulset [upgrade strategy](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#update-strategies) is configured in the `updateStrategy` field. It supports two type of strategies.

- `RollingUpdate`. The statefulset will automatically apply a rolling upgrade to the pods.
- `OnDelete`. The statefulset will not automatically upgrade pods. Pods will be updated when they are recreated after being deleted.

In both cases, the upgrade is initiated when the Pod template is updated.

For BookKeeper, the operator uses an `OnDelete` strategy. With `RollingUpdate` strategy, you can only check the upgrade status once all pods get upgraded. On the other hand, with `OnDelete` you can keep updating pod one by one and keep checking the application status to make sure the upgrade working fine. This allows the operator to have control over the upgrade process and perform verifications and actions before and after a BookKeeper pod is upgraded. For example, checking that there are no under-replicated ledgers before upgrading the next pod. Also, the operator might be need to apply migrations when upgrading to a certain version.

BookKeeper upgrade process is as follows:

1. Statefulset Pod template is updated to the new image and tag according to the Pravega version.
2. Pick one outdated pod
3. Apply pre-upgrade actions and verifications
4. Delete the pod. The pod is recreated with an updated spec and version
5. Wait for the pod to become ready. If it fails to start or times out, the upgrade is cancelled. Check [Recovering from a failed upgrade](#recovering-from-a-failed-upgrade)
6. Apply post-upgrade actions and verifications
7. If all pods are updated, BookKeeper upgrade is completed. Otherwise, go to 2.


### Monitor the upgrade process

You can monitor the upgrade process by listing the Bookkeeper clusters. If a desired version is shown, it means that the operator is working on updating the version.

```
$ kubectl get bk
NAME         VERSION   DESIRED VERSION   DESIRED MEMBERS   READY MEMBERS       AGE
bookkeeper   0.4.0     0.5.0                 4                 3               1h
```

When the upgrade process has finished, the version will be updated.

```
$ kubectl get bk
NAME         VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
bookkeeper   0.5.0     4                 4               1h
```

The command `kubectl describe` can be used to track progress of the upgrade.
```
$ kubectl describe bk bookkeeper
...
Status:
  Conditions:
    Status:                True
    Type:                  Upgrading
    Reason:                Updating BookKeeper
    Message:               1
    Last Transition Time:  2019-04-01T19:42:37+02:00
    Last Update Time:      2019-04-01T19:42:37+02:00
    Status:                False
    Type:                  PodsReady
    Last Transition Time:  2019-04-01T19:43:08+02:00
    Last Update Time:      2019-04-01T19:43:08+02:00
    Status:                False
    Type:                  Error
...  

```
The `Reason` field in Upgrading Condition shows the component currently being upgraded and `Message` field reflects number of successfully upgraded replicas in this component.

If upgrade has failed, please check the `Status` section to understand the reason for failure.

```
$ kubectl describe bk bookkeeper
...
Status:
  Conditions:
    Status:                False
    Type:                  Upgrading
    Last Transition Time:  2019-04-01T19:42:37+02:00
    Last Update Time:      2019-04-01T19:42:37+02:00
    Status:                False
    Type:                  PodsReady
    Last Transition Time:  2019-04-01T19:43:08+02:00
    Last Update Time:      2019-04-01T19:43:08+02:00
    Message:               pod bookkeeper-bookie-0 update failed because of ImagePullBackOff
    Reason:                UpgradeFailed
    Status:                True
    Type:                  Error
  Current Replicas:        8
  Current Version:         0.4.0
  Members:
    Ready:
      bookkeeper-bookie-1
      bookkeeper-bookie-2
      bookkeeper-bookie-3
    Unready:
      bookkeeper-bookie-0
  Ready Replicas:  3
  Replicas:        4
```

You can also find useful information at the operator logs.

```
...
INFO[5884] syncing cluster version from 0.4.0 to 0.5.0-1
INFO[5885] Reconciling BookkeeperCluster default/bookkeeper
INFO[5886] updating statefulset (bookkeeper-bookie) template image to 'pravega/bookkeeper:0.5.0-1'
INFO[5896] Reconciling BookkeeperCluster default/bookkeeper
INFO[5897] statefulset (bookkeeper-bookie) status: 0 updated, 3 ready, 3 target
INFO[5897] updating pod: bookkeeper-bookie-0
INFO[5899] Reconciling BookkeeperCluster default/bookkeeper
INFO[5900] statefulset (bookkeeper-bookie) status: 0 updated, 2 ready, 3 target
INFO[5929] Reconciling BookkeeperCluster default/bookkeeper
INFO[5930] statefulset (bookkeeper-bookie) status: 0 updated, 2 ready, 3 target
INFO[5930] error syncing cluster version, upgrade failed. pod bookkeeper-bookie-0 update failed because of ImagePullBackOff
...
```

### Recovering from a failed upgrade

See [Rollback](rollback-cluster.md)
