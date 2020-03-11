# Bookkeeper Operator

 [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/pravega/bookkeeper-operator?status.svg)](https://godoc.org/github.com/pravega/bookkeeper-operator) [![Build Status](https://travis-ci.org/pravega/bookkeeper-operator.svg?branch=master)](https://travis-ci.org/pravega/bookkeeper-operator) [![Go Report](https://goreportcard.com/badge/github.com/pravega/bookkeeper-operator)](https://goreportcard.com/report/github.com/pravega/bookkeeper-operator) [![Version](https://img.shields.io/github/release/pravega/bookkeeper-operator.svg)](https://github.com/pravega/bookkeeper-operator/releases)

### Project status: alpha

The project is currently alpha. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

## Table of Contents

 * [Overview](#overview)
 * [Requirements](#requirements)
 * [Quickstart](#quickstart)    
    * [Install the Operator](#install-the-operator)
    * [Install a sample Bookkeeper Cluster](#install-a-sample-bookkeeper-cluster)
    * [Scale a Bookkeeper Cluster](#scale-a-bookkeeper-cluster)
    * [Upgrade a Bookkeeper Cluster](#upgrade-a-bookkeeper-cluster)
    * [Uninstall the Bookkeeper Cluster](#uninstall-the-bookkeeper-cluster)
    * [Uninstall the Operator](#uninstall-the-operator)
    * [Manual installation](#manual-installation)
 * [Configuration](#configuration)
 * [Development](#development)
* [Releases](#releases)
* [Upgrade the Operator](#upgrade-the-operator)

## Overview

[Bookkeeper](https://bookkeeper.apache.org/) A scalable, fault-tolerant, and low-latency storage service optimized for real-time workloads.

The Bookkeeper Operator manages Bookkeeper clusters deployed to Kubernetes and automates tasks related to operating a Bookkeeper cluster.

- [x] Create and destroy a Bookkeeper cluster
- [x] Resize cluster
- [x] Rolling upgrades

## Requirements

- Kubernetes 1.9+
- Helm 2.10+
- An existing Apache Zookeeper 3.5 cluster. This can be easily deployed using our [Zookeeper operator](https://github.com/pravega/zookeeper-operator)

## Quickstart

### Install the Operator

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](doc/development.md#installation-on-google-kubernetes-engine).

Use Helm to quickly deploy a Bookkeeper operator with the release name `pravega-bk`.

```
$ helm install charts/bookkeeper-operator --name pr
```

Verify that the Bookkeeper Operator is running.

```
$ kubectl get deploy
NAME                          DESIRED   CURRENT   UP-TO-DATE   AVAILABLE     AGE
pr-bookkeeper-operator          1         1         1            1           17s
```

### Install a sample Bookkeeper cluster

If the BookKeeper cluster is expected to work with Pravega, create a ConfigMap which contains the correct value for the key `PRAVEGA_CLUSTER_NAME`, and provide the name of this file within the field `envVars` present in the BookKeeper Spec. For more details about this ConfigMap refer to [this](doc/bookkeeper-options.md#bookkeeper-custom-configuration).

Helm can be used to install a sample Bookkeeper cluster.

```
$ helm install charts/bookkeeper --name pravega-bk --set zookeeperUri=[ZOOKEEPER_HOST]
```

where:

- `[ZOOKEEPER_HOST]` is the host or IP address of your Zookeeper deployment (e.g. `zk-client:2181`). Multiple Zookeeper URIs can be specified, use a comma-separated list and DO NOT leave any spaces in between (e.g. `zk-0:2181,zk-1:2181,zk-2:2181`).

Check out the [Bookkeeper Helm Chart](charts/bookkeeper) for more a complete list of installation parameters.

Verify that the cluster instances and its components are being created.

```
$ kubectl get bk
NAME                   VERSION   DESIRED MEMBERS   READY MEMBERS      AGE
pravega-bk             0.6.1       3                 1                25s
```

After a couple of minutes, all cluster members should become ready.

```
$ kubectl get bk
NAME                   VERSION   DESIRED MEMBERS   READY MEMBERS     AGE
pravega-bk              0.6.1     3                 3               2m
```

```
$ kubectl get all -l bookkeeper_cluster=pravega-bk
NAME                                              READY   STATUS    RESTARTS   AGE
pod/pravega-bk-bookie-0                              1/1     Running   0          2m
pod/pravega-bk-bookie-1                              1/1     Running   0          2m
pod/pravega-bk-bookie-2                              1/1     Running   0          2m

NAME                                            TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)              AGE
service/pravega-bk-bookie-headless              ClusterIP   None          <none>        3181/TCP             2m

NAME                                            DESIRED   CURRENT     AGE
statefulset.apps/pravega-bk-bookie                 3         3         2m
```

By default, a `BookkeeperCluster` is reachable using this kind of headless service URL for each pod:
```
http://pravega-bk-bookie-0.pravega-bk-bookie-headless.pravega-bk-bookie:3181
```

### Scale a Bookkeeper cluster

You can scale Bookkeeper cluster by updating the `replicas` field in the BookkeeperCluster Spec.

Example of patching the Bookkeeper Cluster resource to scale the server instances to 4.

```
kubectl patch bk pravega-bk --type='json' -p='[{"op": "replace", "path": "/spec/replicas", "value": 4}]'
```

### Upgrade a Bookkeeper cluster

Check out the [upgrade guide](doc/upgrade-cluster.md).

### Uninstall the Bookkeeper cluster

```
$ helm delete pravega-bk --purge
```

### Uninstall the Operator
> Note that the Bookkeeper clusters managed by the Bookkeeper operator will NOT be deleted even if the operator is uninstalled.

```
$ helm delete pr --purge
```
If you want to delete the Bookkeeper cluster, make sure to do it before uninstalling the operator. Also, once the Bookkeeper cluster has been deleted, make sure to check that the zookeeper metadata has been cleaned up before proceeding with the deletion of the operator. This can be confirmed with the presence of the following log message in the operator logs.
```
zookeeper metadata deleted
```

### Manual installation

You can also manually install/uninstall the operator and Bookkeeper with `kubectl` commands. Check out the [manual installation](doc/manual-installation.md) document for instructions.

## Configuration

Check out the [configuration document](doc/configuration.md).

## Development

Check out the [development guide](doc/development.md).

## Releases  

The latest Bookkeeper releases can be found on the [Github Release](https://github.com/pravega/bookkeeper-operator/releases) project page.

## Upgrade the Bookkeeper-Operator
Bookkeeper operator can be upgraded by modifying the image tag using
```
$ kubectl edit <operator deployment name>
```
