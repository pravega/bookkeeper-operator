module github.com/pravega/bookkeeper-operator

go 1.13

require (
	github.com/alecthomas/gocyclo v0.0.0-20150208221726-aa8f8b160214 // indirect
	github.com/hashicorp/go-version v1.1.0
	github.com/mdempsky/unconvert v0.0.0-20200228143138-95ecdbfc0b5f // indirect
	github.com/mibk/dupl v1.0.0 // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/opennota/check v0.0.0-20180911053232-0c771f5545ff // indirect
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/securego/gosec v0.0.0-20200401082031-e946c8c39989 // indirect
	github.com/sirupsen/logrus v1.5.0
	golang.org/x/tools v0.0.0-20200426102838-f3a5411a4c3b // indirect
	k8s.io/api v0.17.5
	k8s.io/apimachinery v0.17.5
	k8s.io/client-go v12.0.0+incompatible
	mvdan.cc/interfacer v0.0.0-20180901003855-c20040233aed // indirect
	mvdan.cc/lint v0.0.0-20170908181259-adc824a0674b // indirect
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
	k8s.io/client-go => k8s.io/client-go v0.17.5
)
