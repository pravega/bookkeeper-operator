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

To install the bookkeeper chart, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/bookkeeper --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_HOST] --set pravegaClusterName=[PRAVEGA_CLUSTER_NAME] -n [NAMESPACE]
```
where:
- **[RELEASE_NAME]** is the release name for the bookkeeper chart
- **[CLUSTER_NAME]** is the name of the bookkeeper cluster so created (if [RELEASE_NAME] contains the string `bookkeeper`, `[CLUSTER_NAME] = [RELEASE_NAME]`, else `[CLUSTER_NAME] = [RELEASE_NAME]-bookkeeper`. The [CLUSTER_NAME] can however be overridden by providing `--set fullnameOverride=[CLUSTER_NAME]` along with the helm install command)
- **[PRAVEGA_CLUSTER_NAME]** is the name of the pravega cluster (this field is optional and needs to be provided only if we expect the bookkeeper cluster to work with [Pravega](https://github.com/pravega/pravega) and if we wish to override its default value which is `pravega`)
- **[VERSION]** can be any stable release version for bookkeeper from 0.5.0 onwards
- **[ZOOKEEPER_HOST]** is the zookeeper service endpoint of your zookeeper cluster deployment (default value of this field is `zookeeper-client:2181`)
- **[NAMESPACE]** is the namespace in which you wish to deploy the bookkeeper cluster (default value for this field is `default`) The bookkeeper cluster must be installed in the same namespace as the zookeeper cluster.

This command deploys bookkeeper on the Kubernetes cluster in its default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

>Note: If the operator version is 0.1.3 or below and bookkeeper version is 0.9.0 or above, you would need to set the JVM options as shown below.
```
helm install [RELEASE_NAME] pravega/bookkeeper --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_HOST] --set pravegaClusterName=[PRAVEGA_CLUSTER_NAME] -n [NAMESPACE] --set 'jvmOptions.extraOpts={-XX:+UseContainerSupport,-XX:+IgnoreUnrecognizedVMOptions}'
```

## Uninstalling the Chart

To uninstall/delete the bookkeeper chart, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

This command removes all the Kubernetes components associated with the chart and deletes the release.
> Note: If blockOwnerDeletion had been set to false during bookkeeper installation, the PVCs won't be removed automatically while uninstalling the bookkeeper chart, and would need to be deleted manually.

## Configuration

The following table lists the configurable parameters of the Bookkeeper chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `version` | Bookkeeper version | `0.7.0` |
| `image.repository` | Image repository | `pravega/bookkeeper` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `replicas` | Number of bookkeeper replicas | `3` |
| `maxUnavailableBookkeeperReplicas` | Number of maxUnavailableBookkeeperReplicas for bookkeeper PDB | `1` |
| `zookeeperUri` | Zookeeper client service URI | `zookeeper-client:2181` |
| `pravegaClusterName` | Name of the pravega cluster | `pravega` |
| `autoRecovery`| Enable bookkeeper auto-recovery | `true` |
| `blockOwnerDeletion`| Enable blockOwnerDeletion | `true` |
| `probes.readiness.initialDelaySeconds` | Number of seconds after the container has started before readiness probe is initiated | `20` |
| `probes.readiness.periodSeconds` | Number of seconds in which readiness probe will be performed | `10` |
| `probes.readiness.failureThreshold` | Number of seconds after which the readiness probe times out | `9` |
| `probes.readiness.successThreshold` | Minimum number of consecutive successes for the readiness probe to be considered successful after having failed | `1` |
| `probes.readiness.timeoutSeconds` | Number of times Kubernetes will retry after a readiness probe failure before restarting the container | `5` |
| `probes.liveness.initialDelaySeconds` | Number of seconds after the container has started before liveness probe is initiated | `60` |
| `probes.liveness.periodSeconds` | Number of seconds in which liveness probe will be performed  | `15` |
| `probes.liveness.failureThreshold` | Number of seconds after which the liveness probe times out | `4` |
| `probes.liveness.successThreshold` | Minimum number of consecutive successes for the liveness probe to be considered successful after having failed | `1` |
| `probes.liveness.timeoutSeconds` | Number of times Kubernetes will retry after a liveness probe failure before restarting the container | `5` |
| `resources.requests.cpu` | Requests for CPU resources | `1000m` |
| `resources.requests.memory` | Requests for memory resources | `4Gi` |
| `resources.limits.cpu` | Limits for CPU resources | `2000m` |
| `resources.limits.memory` | Limits for memory resources | `4Gi` |
| `storage.ledger.className` | Storage class name for bookkeeper ledgers | `` |
| `storage.ledger.volumeSize` | Requested size for bookkeeper ledger persistent volumes | `10Gi` |
| `storage.journal.className` | Storage class name for bookkeeper journals | `` |
| `storage.journal.volumeSize` | Requested size for bookkeeper journal persistent volumes | `10Gi` |
| `storage.index.className` | Storage class name for bookkeeper index | `` |
| `storage.index.volumeSize` | Requested size for bookkeeper index persistent volumes | `10Gi` |
| `jvmOptions.memoryOpts` | Memory Options passed to the JVM for bookkeeper performance tuning | `["-Xms1g", "-XX:MaxDirectMemorySize=2g"]` |
| `jvmOptions.gcOpts` | Garbage Collector (GC) Options passed to the JVM for bookkeeper bookkeeper performance tuning | `[]` |
| `jvmOptions.gcLoggingOpts` | GC Logging Options passed to the JVM for bookkeeper performance tuning | `[]` |
| `jvmOptions.extraOpts` | Extra Options passed to the JVM for bookkeeper performance tuning | `[]` |
| `options` | List of bookkeeper options | |
