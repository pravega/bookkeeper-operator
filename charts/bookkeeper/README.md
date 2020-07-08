# Bookkeeper Helm Chart

Installs Bookkeeper clusters atop Kubernetes.

## Introduction

This chart creates a Bookkeeper cluster in [Kubernetes](http://kubernetes.io) using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Bookkeeper cluster on multiple namespaces.

## Prerequisites

  - Kubernetes 1.15+ with Beta APIs
  - Helm 3+
  - An existing Apache Zookeeper 3.6.1 cluster. This can be easily deployed using our [Zookeeper Operator](https://github.com/pravega/zookeeper-operator)
  - Bookkeeper Operator. You can install it using its own [Helm chart](https://github.com/pravega/bookkeeper-operator/tree/master/charts/bookkeeper-operator)

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install my-release bookkeeper
```

The command deploys bookkeeper on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Bookkeeper chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `version` | Bookkeeper version | `0.7.0` |
| `image.repository` | Image repository | `pravega/bookkeeper` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `replicas` | Number of bookkeeper replicas | `3` |
| `zookeeperUri` | Zookeeper client service URI | `zookeeper-client:2181` |
| `pravegaClusterName` | Name of the pravega cluster | `pravega` |
| `autoRecovery`| Enable bookkeeper auto-recovery | `true` |
| `resources.requests.cpu` | Requests for CPU resources | `1000m` |
| `resources.requests.memory` | Requests for memory resources | `4Gi` |
| `resources.limits.cpu` | Limits for CPU resources | `2000m` |
| `resources.limits.memory` | Limits for memory resources | `4Gi` |
| `storage.ledger.className` | Storage class name for bookkeeper ledgers | `standard` |
| `storage.ledger.volumeSize` | Requested size for bookkeeper ledger persistent volumes | `10Gi` |
| `storage.journal.className` | Storage class name for bookkeeper journals | `standard` |
| `storage.journal.volumeSize` | Requested size for bookkeeper journal persistent volumes | `10Gi` |
| `storage.index.className` | Storage class name for bookkeeper index | `standard` |
| `storage.index.volumeSize` | Requested size for bookkeeper index persistent volumes | `10Gi` |
| `jvmOptions.memoryOpts` | Memory Options passed to the JVM for bookkeeper performance tuning | `["-Xms1g", "-XX:MaxDirectMemorySize=2g"]` |
| `jvmOptions.gcOpts` | Garbage Collector (GC) Options passed to the JVM for bookkeeper bookkeeper performance tuning | `[]` |
| `jvmOptions.gcLoggingOpts` | GC Logging Options passed to the JVM for bookkeeper performance tuning | `[]` |
| `jvmOptions.extraOpts` | Extra Options passed to the JVM for bookkeeper performance tuning | `[]` |
| `options` | List of bookkeeper options | |
