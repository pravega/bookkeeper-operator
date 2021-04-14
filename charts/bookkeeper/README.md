# Deploying Bookkeeper

Here, we briefly describe how to [install](#installing-bookkeeper-cluster)/[update](#updating-bookkeeper-cluster)/[uninstall](#uninstalling-the-bookkeeper-cluster)/[configure](#configuration) bookkeeper clusters atop kubernetes.

## Prerequisites

  - Kubernetes 1.15+ with Beta APIs
  - Helm 3+
  - An existing Apache Zookeeper 3.6.1 cluster. This can be easily deployed using our [Zookeeper Operator](https://github.com/pravega/zookeeper-operator)
  - Bookkeeper Operator. Please refer [this](../../charts/bookkeeper-operator/README.md)

## Installing Bookkeeper Cluster

To install the bookkeeper cluster, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/bookkeeper --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_HOST] --set pravegaClusterName=[PRAVEGA_CLUSTER_NAME] -n [NAMESPACE]
```
where:
- **[RELEASE_NAME]** is the release name for the bookkeeper chart
- **[VERSION]** can be any stable release version for bookkeeper from 0.5.0 onwards
- **[ZOOKEEPER_HOST]** is the zookeeper service endpoint of your zookeeper cluster deployment (default value of this field is `zookeeper-client:2181`)
- **[NAMESPACE]** is the namespace in which you wish to deploy the bookkeeper cluster (default value for this field is `default`) The bookkeeper cluster must be installed in the same namespace as the zookeeper cluster.

>Note: If we provide [RELEASE_NAME] same as chart name, cluster name will be same as release-name. But if we are providing a different name for release(other than bookkeeper in this case), cluster name will be [RELEASE_NAME]-[chart-name]. However, cluster name can be overridden by providing `--set  fullnameOverride=[CLUSTER_NAME]` along with helm install command.

This command deploys bookkeeper on the Kubernetes cluster in its default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

>Note: If the operator version is 0.1.3 or below and bookkeeper version is 0.9.0 or above, you would need to set the JVM options as shown below.
```
helm install [RELEASE_NAME] pravega/bookkeeper --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_HOST] --set pravegaClusterName=[PRAVEGA_CLUSTER_NAME] -n [NAMESPACE] --set 'jvmOptions.extraOpts={-XX:+UseContainerSupport,-XX:+IgnoreUnrecognizedVMOptions}'
```

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
## Updating Bookkeeper Cluster

For updating the bookkeeper cluster, use the following command

```
helm upgrade [RELEASE_NAME]  --version=[VERSION]  --set replicas=5
```
 we can also update other configurable parameters at run time. For changing options `minorCompactionInterval` to `1900` use the below command.

 ```
  helm upgrade bookkeeper charts/bookkeeper --set-string options."minorCompactionInterval=1900"
  ```
Please refer [upgrade](../../doc/upgrade-cluster.md) for upgrading cluster versions.

## Uninstalling the Bookkeeper cluster

To uninstall/delete the bookkeeper cluster, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

This command removes all the Kubernetes components associated with the chart and deletes the release.
> Note: If blockOwnerDeletion had been set to false during bookkeeper installation, the PVCs won't be removed automatically while uninstalling the bookkeeper chart, and would need to be deleted manually.

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

## Configuration

The following table lists the configurable parameters of the Bookkeeper chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `version` | Bookkeeper version | `0.9.0` |
| `image.repository` | Image repository | `pravega/bookkeeper` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `replicas` | Number of bookkeeper replicas | `3` |
| `maxUnavailableReplicas` | Maximum number of unavailable replicas for bookkeeper PDB | |
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
| `jvmOptions.gcLoggingOpts` | GC Logging Options passed to the JVM for bookkeeper performance tuning | `["-Xlog:gc*,safepoint::time,level,tags:filecount=5,filesize=64m"]` |
| `jvmOptions.extraOpts` | Extra Options passed to the JVM for bookkeeper performance tuning | `["-XX:+IgnoreUnrecognizedVMOptions"]` |
| `options` | List of bookkeeper options | |
| `labels` | Labels to be added to the Bookie Pods | |
