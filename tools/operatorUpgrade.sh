#! /bin/bash
set -ex

echo "Running pre-upgrade script for upgrading bookeeper operator from version prior to 0.1.3 to 0.1.3"

if [ "$#" -ne 3 ]; then
	echo "Error : Invalid number of arguments"
	Usage: "./operatorUpgrade.sh <bookkeeper-operator deployment name> <bookkeeper-operator deployment namespace> <bookkeeper-operator new image-repo:image-tag>"
	exit 1
fi

function UpgradingToBookkeeperoperator(){

local op_deployment_name=$1

local namespace=$2

local op_name=`kubectl describe deploy ${op_deployment_name} -n ${namespace}| grep "Name:" | awk '{print $2}' | head -1`

local op_image=$3

sed -i "s/namespace.*/namespace: $namespace"/ ./manifest_files/certificate.yaml

local temp_string_for_dns=bookkeeper-webhook-svc.${namespace}

sed -i "s/bookkeeper-webhook-svc.default/${temp_string_for_dns}"/ ./manifest_files/certificate.yaml

#Installing the certificate
kubectl apply -f  ./manifest_files/certificate.yaml

#Reverting the changes back in the certificate.yaml file
sed -i "s/${temp_string_for_dns}/pravega-webhook-svc.default"/ ./manifest_files/certificate.yaml

sed -i "s|cert.*|cert-manager.io/inject-ca-from: $namespace/selfsigned-cert-bk|" ./manifest_files/webhook.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/webhook.yaml

#Insalling the webhook
kubectl apply -f ./manifest_files/webhook.yaml

sed -i "s|image:.*|image: $op_image|" ./manifest_files/patch.yaml

sed -i "s/value:.*/value: $op_name "/ ./manifest_files/patch.yaml

sed -i "/imagePullPolicy:.*/{n;s/name.*/name: $op_name/}" ./manifest_files/patch.yaml

#deleting the mutatingwebhookconfiguration created by the previous operator
kubectl delete mutatingwebhookconfiguration bookkeeper-webhook-config

#updating the operator using patch file
kubectl patch deployment $op_name --namespace ${namespace} --type merge --patch "$(cat ./manifest_files/patch.yaml)"

}

UpgradingToBookkeeperoperator $1 $2 $3
