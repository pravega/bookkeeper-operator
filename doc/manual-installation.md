## Manual installation

* [Install the Operator manually](#install-the-operator-manually)
* [Install the Pravega cluster manually](#install-the-bookkeeper-cluster-manually)
* [Uninstall the Pravega Cluster manually](#uninstall-the-bookkeeper-cluster-manually)
* [Uninstall the Operator manually](#uninstall-the-operator-manually)

### Install the Operator manually

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](#installation-on-google-kubernetes-engine).

Register the Pravega cluster custom resource definition (CRD).

```
$ kubectl create -f deploy/crds/crd.yaml
```

Create the operator role, role binding and service account.

```
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
$ kubectl create -f deploy/service_account.yaml
```

Install the operator.

```
$ kubectl create -f deploy/operator.yaml
```

### Install the Bookkeeper cluster manually

Once the operator is installed, you can use the following YAML template to install a small development Bookkeeper Cluster. Create a `bookkeeper.yaml` file with the following content.

```yaml
apiVersion: "bookkeeper.pravega.io/v1alpha1"
kind: "BookkeeperCluster"
metadata:
  name: "pravega-bk"
spec:
  version: 0.6.1
  zookeeperUri: [ZOOKEEPER_HOST]:2181

  bookkeeper:
    replicas: 3
    image:
      repository: pravega/bookkeeper
    autoRecovery: true

```

where:

- `[ZOOKEEPER_HOST]` is the host or IP address of your Zookeeper deployment.

Check out other sample CR files in the [`example`](../example) directory.

Deploy the Bookkeeper cluster.

```
$ kubectl create -f bookkeeper.yaml
```

Verify that the cluster instances and its components are being created.

```
$ kubectl get PravegaCluster
NAME         VERSION   DESIRED MEMBERS    READY MEMBERS      AGE
pravega-bk   0.6.1      3                    0               25s
```

### Uninstall the Bookkeeper cluster manually

```
$ kubectl delete -f bookkeeper.yaml
```

### Uninstall the Operator manually

> Note that the Bookkeeper cluster managed by the Bookkeeper operator will NOT be deleted even if the operator is uninstalled.

To delete all clusters, delete all cluster CR objects before uninstalling the operator.

```
$ kubectl delete -f deploy
```
