# Bookkeeper Operator

 [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/pravega/bookkeeper-operator?status.svg)](https://godoc.org/github.com/pravega/bookkeeper-operator) [![Build Status](https://travis-ci.org/pravega/bookkeeper-operator.svg?branch=master)](https://travis-ci.org/pravega/bookkeeper-operator) [![Go Report](https://goreportcard.com/badge/github.com/pravega/bookkeeper-operator)](https://goreportcard.com/report/github.com/pravega/bookkeeper-operator) [![Version](https://img.shields.io/github/release/pravega/bookkeeper-operator.svg)](https://github.com/pravega/bookkeeper-operator/releases)

### Project status: alpha

The project is currently alpha. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

## Table of Contents

 * [Overview](#overview)
 * [Requirements](#requirements)
 * [Quickstart](#quickstart)    
    * [Install the Operator](#install-the-operator)
        * [Install the Operator in Test Mode](#install-the-operator-in-test-mode)
    * [Install a sample Bookkeeper Cluster](#install-a-sample-bookkeeper-cluster)
    * [Scale a Bookkeeper Cluster](#scale-a-bookkeeper-cluster)
    * [Upgrade a Bookkeeper Cluster](#upgrade-a-bookkeeper-cluster)
    * [Upgrade the Operator](#upgrade-the-operator)
    * [Uninstall the Bookkeeper Cluster](#uninstall-the-bookkeeper-cluster)
    * [Uninstall the Operator](#uninstall-the-operator)
    * [Manual installation](#manual-installation)
 * [Configuration](#configuration)
 * [Development](#development)
* [Releases](#releases)

## Overview

[Bookkeeper](https://bookkeeper.apache.org/) is a scalable, fault-tolerant, and low-latency storage service optimized for real-time workloads.

The Bookkeeper Operator manages Bookkeeper clusters deployed to Kubernetes and automates tasks related to operating a Bookkeeper cluster.

- [x] Create and destroy a Bookkeeper cluster
- [x] Resize cluster
- [x] Rolling upgrades

## Requirements

- Kubernetes 1.15+
- Helm 3.2.1+
- An existing Apache Zookeeper 3.6.1 cluster. This can be easily deployed using our [Zookeeper operator](https://github.com/pravega/zookeeper-operator)

## Quickstart

We recommend using our [helm charts](charts) for all installation and upgrades (but not for rollbacks at the moment since helm rollbacks are still experimental). The helm charts for bookkeeper operator (version 0.1.2 onwards) and bookkeeper cluster (version 0.5.0 onwards) are published in [https://charts.pravega.io](https://charts.pravega.io/). To add this repository to your Helm repos, use the following command:
```
helm repo add pravega https://charts.pravega.io
```
There are manual deployment, upgrade and rollback options available as well.

### Install the Operator

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](doc/development.md#installation-on-google-kubernetes-engine).

To understand how to deploy a Bookkeeper Operator using helm, refer to [this](charts/bookkeeper-operator#installing-the-chart).

#### Install the Operator in Test Mode
 The Operator can be run in `test mode` if we want to deploy the Bookkeeper Cluster on minikube or on a cluster with very limited resources by setting `testmode: true` in `values.yaml` file. Operator running in test mode skips the minimum replica requirement checks. Test mode provides a bare minimum setup and is not recommended to be used in production environments.

### Install a sample Bookkeeper cluster

To understand how to deploy a bookkeeper cluster using helm, refer to [this](charts/bookkeeper#installing-the-chart).

Once the bookkeeper cluster with release name `bookkeeper` has been created, use the following command to verify that the cluster instances and its components are being created.

```
$ kubectl get bk
NAME                   VERSION   DESIRED MEMBERS   READY MEMBERS      AGE
bookkeeper             0.7.0     3                 1                  25s
```

After a couple of minutes, all cluster members should become ready.

```
$ kubectl get bk
NAME                   VERSION   DESIRED MEMBERS   READY MEMBERS     AGE
bookkeeper             0.7.0     3                 3                 2m
```

```
$ kubectl get all -l bookkeeper_cluster=bookkeeper
NAME                                              READY   STATUS    RESTARTS   AGE
pod/bookkeeper-bookie-0                           1/1     Running   0          2m
pod/bookkeeper-bookie-1                           1/1     Running   0          2m
pod/bookkeeper-bookie-2                           1/1     Running   0          2m

NAME                                            TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)              AGE
service/bookkeeper-bookie-headless              ClusterIP   None          <none>        3181/TCP             2m

NAME                                            DESIRED   CURRENT     AGE
statefulset.apps/bookkeeper-bookie              3         3           2m
```

By default, a `BookkeeperCluster` is reachable using this kind of headless service URL for each pod:
```
http://bookkeeper-bookie-0.bookkeeper-bookie-headless.bookkeeper-bookie:3181
```

### Scale a Bookkeeper cluster

You can scale Bookkeeper cluster by updating the `replicas` field in the BookkeeperCluster Spec.

Example of patching the Bookkeeper Cluster resource to scale the server instances to 4.

```
kubectl patch bk bookkeeper --type='json' -p='[{"op": "replace", "path": "/spec/replicas", "value": 4}]'
```

### Upgrade a Bookkeeper cluster

Check out the [upgrade guide](doc/upgrade-cluster.md).

## Upgrade the Operator

For upgrading the bookkeeper operator check the document [operator-upgrade](doc/operator-upgrade.md)

### Uninstall the Bookkeeper cluster

```
$ helm uninstall [BOOKKEEPER_RELEASE_NAME]
```

Once the Bookkeeper cluster has been deleted, make sure to check that the zookeeper metadata has been cleaned up before proceeding with the deletion of the operator. This can be confirmed with the presence of the following log message in the operator logs.
```
zookeeper metadata deleted
```

However, if the operator fails to delete this metadata from zookeeper, you will instead find the following log message in the operator logs.
```
failed to cleanup [CLUSTER_NAME] metadata from zookeeper (znode path: /pravega/[PRAVEGA_CLUSTER_NAME]): <error-msg>
```

The operator additionally sends out a `ZKMETA_CLEANUP_ERROR` event to notify the user about this failure. The user can check this event by doing `kubectl get events`. The following is the sample describe output of the event that is generated by the operator in such a case
```
Name:             ZKMETA_CLEANUP_ERROR-nn6sd
Namespace:        default
Labels:           app=bookkeeper-cluster
                  bookkeeper_cluster=bookkeeper
Annotations:      <none>
API Version:      v1
Event Time:       <nil>
First Timestamp:  2020-04-27T16:53:34Z
Involved Object:
  API Version:   app.k8s.io/v1beta1
  Kind:          Application
  Name:          bookkeeper-cluster
  Namespace:     default
Kind:            Event
Last Timestamp:  2020-04-27T16:53:34Z
Message:         failed to cleanup bookkeeper metadata from zookeeper (znode path: /pravega/pravega): failed to delete zookeeper znodes for (bookkeeper): failed to connect to zookeeper: lookup zookeeper-client on 10.100.200.2:53: no such host
Metadata:
  Creation Timestamp:  2020-04-27T16:53:34Z
  Generate Name:       ZKMETA_CLEANUP_ERROR-
  Resource Version:    864342
  Self Link:           /api/v1/namespaces/default/events/ZKMETA_CLEANUP_ERROR-nn6sd
  UID:                 5b4c3f80-36b5-43e6-b417-7992bc309218
Reason:                ZK Metadata Cleanup Failed
Reporting Component:   bookkeeper-operator
Reporting Instance:    bookkeeper-operator-6769886978-xsjx6
Source:
Type:    Error
Events:  <none>
```

>In case the operator fails to delete the zookeeper metadata, the user is expected to manually delete the metadata from zookeeper prior to reinstall.

### Uninstall the Operator
> Note that the Bookkeeper clusters managed by the Bookkeeper operator will NOT be deleted even if the operator is uninstalled.

```
$ helm uninstall [BOOKKEEPER_OPERATOR_RELEASE_NAME]
```

### Manual installation

You can also manually install/uninstall the operator and Bookkeeper with `kubectl` commands. Check out the [manual installation](doc/manual-installation.md) document for instructions.

## Configuration

Check out the [configuration document](doc/configuration.md).

## Development

Check out the [development guide](doc/development.md).

## Releases  

The latest Bookkeeper releases can be found on the [Github Release](https://github.com/pravega/bookkeeper-operator/releases) project page.
