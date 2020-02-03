## Setting up RBAC for Bookkeeper operator

### Use non-default service accounts

You can optionally configure non-default service accounts for the Bookkeeper.

For BookKeeper, set the `serviceAccountName` field under the `spec` block.

```
...
spec:
    serviceAccountName: bk-service-account
...
```

Replace the `namespace` with your own namespace.

### Installing on a Custom Namespace with RBAC enabled

Create the namespace.

```
$ kubectl create namespace pravega-io
```

Update the namespace configured in the `deploy/role_binding.yaml` file.

```
$ sed -i -e 's/namespace: default/namespace: pravega-io/g' deploy/role_binding.yaml
```

Apply the changes.

```
$ kubectl -n pravega-io apply -f deploy
```

Note that the Pravega operator only monitors the `BookkeeperCluster` resources which are created in the same namespace, `pravega-io` in this example. Therefore, before creating a `BookkeeperCluster` resource, make sure an operator exists in that namespace.

```
$ kubectl -n pravega-io create -f example/cr.yaml
```

```
$ kubectl -n pravega-io get bookkeeperclusters
NAME      AGE
pravega   28m
```

```
$ kubectl -n pravega-io get pods -l pravega_cluster=pravega
NAME                                          READY     STATUS    RESTARTS   AGE
pravega-bookie-0                              1/1       Running   0          29m
pravega-bookie-1                              1/1       Running   0          29m
pravega-bookie-2                              1/1       Running   0          29m
```
