# Upgrade Guide

## Upgrading till 0.1.2

Bookkeeper operator can be upgraded to a version **[VERSION]** via helm using the following command

```
$ helm upgrade [BOOKKEEPER_OPERATOR_RELEASE_NAME] pravega/bookkeeper-operator --version=[VERSION]
```
The bookkeeper operator with deployment name **[DEPLOYMENT_NAME]** can also be upgraded manually by modifying the image tag using kubectl edit, patch or apply
```
$ kubectl edit deploy [DEPLOYMENT_NAME]
```

## Upgrading to 0.1.3

### Pre-requisites

For upgrading Operator to version 0.1.3, the following must be true:
1. The Kubernetes Server version must be at least 1.15, with Beta APIs

2. Cert-Manager v0.15.0+ or some other certificate management solution must be deployed for managing webhook service certificates. The upgrade trigger script assumes that the user has [cert-manager](https://cert-manager.io/docs/installation/kubernetes/) installed but any other cert management solution can also be used and script would need to be modified accordingly.
To install cert-manager check [this](https://cert-manager.io/docs/installation/kubernetes/).

3. Install an Issuer and a Certificate (either self-signed or CA signed) in the same namespace as the Bookkeeper Operator (refer to [this](https://github.com/pravega/bookkeeper-operator/blob/master/deploy/certificate.yaml) manifest to create a self-signed certificate in the default namespace).

4. Execute the script `pre-upgrade.sh` inside the [scripts](https://github.com/pravega/bookkeeper-operator/blob/master/scripts) folder. This script patches the `bookkeeper-webhook-svc` with the required annotations and labels. The format of the command is
```
./pre-upgrade.sh [BOOKKEEPER_OPERATOR_RELEASE_NAME][BOOKKEEPER_OPERATOR_NAMESPACE]
```
where:
- `[BOOKKEEPER_OPERATOR_RELEASE_NAME]` is the release name of the bookkeeper operator deployment
- `[BOOKKEEPER_OPERATOR_NAMESPACE]` is the namespace in which the bookkeeper operator has been deployed (this is an optional parameter and its default value is `default`)

### Triggering the upgrade

#### Upgrade via helm

The upgrade to Operator 0.1.3 can be triggered using the following command
```
helm upgrade [BOOKKEEPER_OPERATOR_RELEASE_NAME] pravega/bookkeeper-operator --version=0.1.3 --set webhookCert.certName=[CERT_NAME] --set webhookCert.secretName=[SECRET_NAME]
```
where:
- `[CERT_NAME]` is the name of the certificate that has been created
- `[SECRET_NAME]` is the name of the secret created by the above certificate

#### Upgrade manually

To manually trigger the upgrade to Operator 0.1.3, run the script `operatorUpgrade.sh` under [tools](https://github.com/pravega/bookkeeper-operator/blob/master/tools) folder. This script installs the certificate, patches and creates necessary K8s artifacts, needed by 0.1.3 Operator, prior to triggering the upgrade by updating the image tag in Operator deployment.
