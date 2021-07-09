# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail

PROJECT_NAME=bookkeeper-operator
REPO=pravega/$(PROJECT_NAME)
BASE_VERSION=0.1.6
ID=$(shell git rev-list `git rev-list --tags --no-walk --max-count=1`..HEAD --count)
GIT_SHA=$(shell git rev-parse --short HEAD)
VERSION=$(BASE_VERSION)-$(ID)-$(GIT_SHA)
GOOS=linux
GOARCH=amd64
TEST_REPO=testbkop/$(PROJECT_NAME)
DOCKER_TEST_PASS=testbkop@123
DOCKER_TEST_USER=testbkop
TEST_IMAGE=$(TEST_REPO)-testimages:$(VERSION)

.PHONY: all build check clean test

all: check build test

build: build-go build-image

build-go:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
	-o bin/$(PROJECT_NAME) cmd/manager/main.go

build-image:
	echo "$(REPO)"
	docker build --no-cache --build-arg VERSION=$(VERSION) --build-arg DOCKER_REGISTRY=$(DOCKER_REGISTRY) --build-arg GIT_SHA=$(GIT_SHA) -t $(REPO):$(VERSION) .
	docker tag $(REPO):$(VERSION) $(REPO):latest

test: test-unit test-e2e

test-unit:
	go test $$(go list ./... | grep -v /vendor/ | grep -v /test/e2e ) -race -coverprofile=coverage.txt -covermode=atomic

test-e2e: test-e2e-remote

test-e2e-remote: login
		operator-sdk build $(TEST_IMAGE)
		docker push $(TEST_IMAGE)
		operator-sdk test local ./test/e2e --operator-namespace default --image $(TEST_IMAGE) --go-test-flags "-v -timeout 0"

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
