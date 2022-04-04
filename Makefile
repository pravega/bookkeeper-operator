# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail
CRD_OPTIONS ?= "crd:trivialVersions=true"

PROJECT_NAME=bookkeeper-operator
REPO=pravega/$(PROJECT_NAME)
BASE_VERSION=0.1.8
ID=$(shell git rev-list HEAD --count)
GIT_SHA=$(shell git rev-parse --short HEAD)
VERSION=$(BASE_VERSION)-$(ID)-$(GIT_SHA)
GOOS=linux
GOARCH=amd64
TEST_REPO=testbkop/$(PROJECT_NAME)
DOCKER_TEST_PASS=testbkop@123
DOCKER_TEST_USER=testbkop
TEST_IMAGE=$(TEST_REPO)-testimages:$(VERSION)

.PHONY: all build check clean test
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: check build test

build: build-go build-image

build-go:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
	-o bin/$(PROJECT_NAME) main.go

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif


build-image:
	echo "$(REPO)"
	docker build --no-cache --build-arg VERSION=$(VERSION) --build-arg DOCKER_REGISTRY=$(DOCKER_REGISTRY) --build-arg GIT_SHA=$(GIT_SHA) -t $(REPO):$(VERSION) .
	docker tag $(REPO):$(VERSION) $(REPO):latest

test: test-unit test-e2e

test-unit:
	go test $$(go list ./... | grep -v /vendor/ | grep -v /test/e2e ) -race -coverprofile=coverage.txt -covermode=atomic

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image pravega/bookkeeper-operator=$(TEST_IMAGE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -


# Undeploy controller in the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

test-e2e: test-e2e-remote

test-e2e-remote:
	make login
	docker build . -t $(TEST_IMAGE)
	docker push $(TEST_IMAGE)
	make deploy
	RUN_LOCAL=false go test -v -timeout 2h ./test/e2e...
	make undeploy

login:
		echo "$(DOCKER_TEST_PASS)" | docker login -u "$(DOCKER_TEST_USER)" --password-stdin

test-e2e-local:
	operator-sdk test local ./test/e2e --namespace default --up-local --go-test-flags "-v -timeout 0"

run-local:
	operator-sdk up local --operator-flags -webhook=false

login:
	echo "$(DOCKER_TEST_PASS)" | docker login -u "$(DOCKER_TEST_USER)" --password-stdin

push: build login
	docker push $(REPO):$(VERSION)
	if [[ ${TRAVIS_TAG} =~ ^([0-9]+\.[0-9]+\.[0-9]+)$$ ]]; then docker push $(REPO):latest; fi;

clean:
	rm -f bin/$(PROJECT_NAME)

check: check-format check-license

check-format:
	./scripts/check_format.sh

check-license:
	./scripts/check_license.sh
