module github.com/pravega/bookkeeper-operator

go 1.16

require (
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/hashicorp/go-version v1.1.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/operator-framework/operator-lib v0.2.0
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.3 // indirect
	golang.org/x/tools v0.1.2 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.0
)

replace (
	#github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
	#github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.12.0
	#github.com/onsi/gomega => github.com/onsi/gomega v1.9.0
	#k8s.io/api => k8s.io/api v0.19.13
	#k8s.io/apimachinery => k8s.io/apimachinery v0.19.14-rc.0
	#k8s.io/client-go => k8s.io/client-go v0.19.13
	#sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.5
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
)
