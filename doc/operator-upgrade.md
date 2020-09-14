# Upgrade Guide

## Upgrading till 0.1.2

Bookkeeper operator can be upgraded via helm using the following command
```
$ helm upgrade bookkeeper-operator <location of modified operator charts>
```
Here `bookkeeper-operator` is the release name of the operator. It can also be upgraded manually by modifying the image tag using the following command
```
$ kubectl edit deploy bookkeeper-operator
```
## Upgrading to 0.1.3

### Pre-requisites

For upgrading Operator to version 0.1.3, the following must be true:
1. The Kubernetes Server version must be at least 1.15, with Beta APIs

2. Cert-Manager v0.15.0+ or some other certificate management solution must be deployed for managing webhook service certificates. The upgrade trigger script assumes that the user has [cert-manager](https://cert-manager.io/docs/installation/kubernetes/) installed but any other cert management solution can also be used and script would need to be modified accordingly.
To install cert-manager check [this](https://cert-manager.io/docs/installation/kubernetes/).

3. Install an Issuer and a Certificate (either self-signed or CA signed) in the same namespace as the Pravega Operator (refer to [this](https://github.com/pravega/bookkeeper-operator/blob/master/deploy/certificate.yaml) manifest to create a self-signed certificate in the default namespace).
> The name of the certificate (*webhookCert.certName*), the name of the secret created by this certificate (*webhookCert.secretName*), the tls.crt (*webhookCert.crt*) and tls.key (*webhookCert.key*) need to be specified against the corresponding fields in the values.yaml file, or can be provided with the upgrade command as shown [here](#triggering-the-upgrade).
The values *tls.crt* and *tls.key* are contained in the secret which is created by the certificate and can be obtained using the following command
```
kubectl get secret <secret-name> -o yaml | grep tls.
```

5. Execute the script `pre-upgrade.sh` inside the [scripts](https://github.com/pravega/bookkeeper-operator/blob/master/scripts) folder. This script patches the `bookkeeper-webhook-svc` with the required annotations and labels.


### Triggering the upgrade

#### Upgrade via helm

The upgrade to Operator 0.1.3 can be triggered using the following command
```
helm upgrade <operator release name> <location of 0.1.3 charts>  --set webhookCert.generate=false --set webhookCert.certName=<cert-name> --set webhookCert.secretName=<secret-name>
```
#### Upgrade manually

To manually trigger the upgrade to Operator 0.1.3, run the script `operatorUpgrade.sh` under [tools](https://github.com/pravega/bookkeeper-operator/blob/master/tools) folder. This script installs certificate, patches and creates necessary K8s artifacts, needed by 0.1.3 Operator, prior to triggering the upgrade by updating the image tag in Operator deployment.
