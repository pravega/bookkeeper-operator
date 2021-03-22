# Troubleshooting

## Bookkeeper Cluster Issues

* [Certificate Error: Internal error occurred: failed calling webhook](#certificate-error-internal-error-occurred-failed-calling-webhook)
* [Unsupported Bookkeeper cluster version](#unsupported-bookkeeper-cluster-version)
* [Unsupported upgrade from version](#unsupported-upgrade-from-version)
* [failed to delete znode](#failed-to-delete-znode)

## Bookkeeper operator Issues
* [Operator pod in container creating state](#operator-pod-in-container-creating-state)

## Certificate Error: Internal error occurred: failed calling webhook

while installing bookkeeper, if we get the error as  below,
```
helm install bookkeeper charts/bookkeeper
Error: Post https://bookkeeper-webhook-svc.default.svc:443/validate-bookkeeper-pravega-io-v1alpha1-bookkeepercluster?timeout=30s: x509: certificate signed by unknown authority
```
We need to ensure that certificates are installed before installing the operator. Please refer [prerequisite](../charts/bookkeeper-operator/README.md#Prerequisites)

## Unsupported Bookkeeper cluster version

while installing pravega, if we get the below error
```
Error: admission webhook "bookkeeperwebhook.pravega.io" denied the request: unsupported Bookkeeper cluster version 0.10.0-2703.c9b7be114
```
We need to make sure the supported versions are present in config map by the following command

`kubectl describe cm bk-supported-versions-map`

If the entries are not there in configmap, we have to add these options in the configmap by enabling test mode as follows while installing operator

```
helm install pravega-operator charts/bookkeeper-operator --set testmode.enabled=true --set testmode.version="0.10.0"
```

Alternatively, we can edit the configmap and add entry as `0.10.0:0.10.0` in the configmap and restart the bookkeeper-operator pod

## Unsupported upgrade from version

while upgrading bookkeeper, if we get the error similar to below

```
Error from server (unsupported upgrade from version 0.8.0-2640.e4c436ba9 to 0.9.0-2752.2652549b3): error when applying patch
```
We need to make sure that supported versions are present in configmap as `0.8.0:0.9.0`. If the entries are missing, we have to add these options in the configmap by enabling test mode as follows while installing Operator

```
helm install bookkeeper-operator charts/bookkeeper-operator --set testmode.enabled=true --set testmode.fromVersion="0.8.0" --set testmode.version="0.9.0"
```
Alternatively, we can edit the configmap and add entry as `0.8.0:0.8.0,0.9.0` in the configmap and restart the pravega-operator pod

## failed to delete znode

while installing bookkeeper, if the pods are not coming to ready state `1/1` and in the operator logs if the error messages are seen as below,

```
time="2021-03-17T11:43:03Z" level=info msg="failed to reconcile bookkeeper cluster (nautilus): failed to clean up zookeeper: failed to cleanup nautilus metadata from zookeeper (znode path: /pravega/nautilus): failed to delete zookeeper znodes for (nautilus): failed to delete znode (/pravega/nautilus/watermarks/ownership): zk: node has children"
```

we need to ensure that znode entries are cleaned up from previous installation. This can be done by either cleaning up znode entries from zookeeper nodes or by completely reinstalling zookeeper.

## Operator pod in container creating state

while installing operator, if the operator pod goes in `ContainerCreating` state for long time, make sure certificates are installed correctly.Please refer [prerequisite](../charts/bookkeeper-operator/README.md#Prerequisites)
