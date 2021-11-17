module github.com/pravega/bookkeeper-operator

go 1.16

require (
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/hashicorp/go-version v1.1.0
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/operator-framework/operator-lib v0.2.0
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.3 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/tools v0.1.2 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.19.13
	k8s.io/apiextensions-apiserver v0.19.0-alpha.1 // indirect
	k8s.io/apimachinery v0.19.13
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.8.0 // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/controller-runtime v0.9.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega => github.com/onsi/gomega v1.9.0
	k8s.io/api => k8s.io/api v0.19.13
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.14-rc.0
	k8s.io/client-go => k8s.io/client-go v0.19.13
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.5
)
