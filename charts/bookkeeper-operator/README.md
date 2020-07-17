# Bookkeeper Operator Helm Chart

Installs [Bookkeeper Operator](https://github.com/pravega/bookkeeper-operator) to create/configure/manage Bookkeeper clusters atop Kubernetes.

## Introduction

This chart bootstraps a [Bookkeeper Operator](https://github.com/pravega/bookkeeper-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Bookkeeper Operator on multiple namespaces.

## Prerequisites
  - Kubernetes 1.15+ with Beta APIs
  - Helm 3+
  - An existing Apache Zookeeper 3.6.1 cluster. This can be easily deployed using our [Zookeeper Operator](https://github.com/pravega/zookeeper-operator)

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install my-release bookkeeper-operator
```

The command deploys bookkeeper operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Bookkeeper operator chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `image.repository` | Image repository | `pravega/bookkeeper-operator` |
| `image.tag` | Image tag | `0.1.2` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `crd.create` | Create bookkeeper CRD | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Name for the service account | `bookkeeper-operator` |
| `testmode.enabled` | Enable test mode | `false` |
| `testmode.version` | Major version number of the alternate bookkeeper image we want the operator to deploy, if test mode is enabled | `""` |
| `watchNamespace` | Namespaces to be watched  | `""` |
