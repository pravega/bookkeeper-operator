# Bookkeeper Operator Helm Chart

Installs [Bookkeeper Operator](https://github.com/pravega/bookkeeper-operator) to create/configure/manage Bookkeeper clusters atop Kubernetes.

## Introduction

This chart bootstraps a [Bookkeeper Operator](https://github.com/pravega/bookkeeper-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Bookkeeper Operator on multiple namespaces.

## Prerequisites
  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+
  - An existing Apache Zookeeper 3.6.1 cluster. This can be easily deployed using our [Zookeeper Operator](https://github.com/pravega/zookeeper-operator)
  - Cert-Manager v0.15.0+ or some other certificate management solution in order to manage the webhook service certificates. This can be easily deployed by referring to [this](https://cert-manager.io/docs/installation/kubernetes/)
  - An Issuer and a Certificate (either self-signed or CA signed) in the same namespace that the Bookkeeper Operator will be installed (refer to [this](https://github.com/pravega/bookkeeper-operator/blob/master/deploy/certificate.yaml) manifest to create a self-signed certificate in the default namespace)

## Installing the Chart

To install the bookkeeper-operator chart, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/bookkeeper-operator --version=[VERSION] --set webhookCert.certName=[CERT_NAME] --set webhookCert.secretName=[SECRET_NAME]
```
where:
- **[RELEASE_NAME]** is the release name for the bookkeeper-operator chart
- **[DEPLOYMENT_NAME]** is the name of the bookkeeper-operator deployment so created. (If [RELEASE_NAME] contains the string `bookkeeper-operator`, `[DEPLOYMENT_NAME] = [RELEASE_NAME]`, else `[DEPLOYMENT_NAME] = [RELEASE_NAME]-bookkeeper-operator`. The [DEPLOYMENT_NAME] can however be overridden by providing `--set fullnameOverride=[DEPLOYMENT_NAME]` along with the helm install command)
- **[VERSION]** can be any stable release version for bookkeeper-operator from 0.1.3 onwards
- **[CERT_NAME]** is the name of the certificate created as a prerequisite
- **[SECRET_NAME]** is the name of the secret created by the above certificate

This command deploys a bookkeeper-operator on the Kubernetes cluster in its default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

>Note: If the bookkeeper-operator version is 0.1.2, webhookCert.certName and webhookCert.secretName should not be set. Also in this case, cert-manager and the certificate/issuer do not need to be deployed as prerequisites.

## Uninstalling the Chart

To uninstall/delete the bookkeeper-operator chart, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the bookkeeper-operator chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `image.repository` | Image repository | `pravega/bookkeeper-operator` |
| `image.tag` | Image tag | `0.1.3` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `crd.create` | Create bookkeeper CRD | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Name for the service account | `bookkeeper-operator` |
| `testmode.enabled` | Enable test mode | `false` |
| `testmode.version` | Major version number of the alternate bookkeeper image we want the operator to deploy or provide an upgrade path to, if test mode is enabled | `""` |
| `testmode.fromVersion` | Major version number of the alternate bookkeeper image, if we wish to provide an upgrade path from this version to the version mentioned above, if test mode is enabled | `""` |
| `webhookCert.crt` | tls.crt value corresponding to the certificate | |
| `webhookCert.key` | tls.key value corresponding to the certificate | |
| `webhookCert.generate` | Whether to generate the certificate and the issuer (set to false while using self-signed certificates) | `false` |
| `webhookCert.certName` | Name of the certificate, if generate is set to false | `selfsigned-cert-bk` |
| `webhookCert.secretName` | Name of the secret created by the certificate, if generate is set to false | `selfsigned-cert-tls-bk` |
| `watchNamespace` | Namespaces to be watched  | `""` |
