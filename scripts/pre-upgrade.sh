#! /bin/bash
set -ex

if [[ "$#" -lt 1 || "$#" -gt 2 ]]; then
        echo "Error : Invalid number of arguments"
        Usage: "./pre-upgrade.sh <bookkeeper-operator-release-name> <bookkeeper-operator-namespace>"
        exit 1
fi

name=$1
namespace=${2:-default}

kubectl annotate Service bookkeeper-webhook-svc meta.helm.sh/release-name=$name -n $namespace --overwrite
kubectl annotate Service bookkeeper-webhook-svc meta.helm.sh/release-namespace=$namespace -n $namespace --overwrite
kubectl label Service bookkeeper-webhook-svc app.kubernetes.io/managed-by=Helm -n $namespace --overwrite

#deleting the mutatingwebhookconfiguration created by the previous operator
kubectl delete mutatingwebhookconfiguration bookkeeper-webhook-config
