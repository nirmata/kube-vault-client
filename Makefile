.DEFAULT_GOAL: build

GIT_VERSION := $(shell git describe --dirty --always --tags)
TAG?=$(GIT_VERSION)
IMAGE?=kube-vault-client
PREFIX?=nirmata/$(IMAGE)
ARCH?=amd64


version:
	@echo "building $(PREFIX):$(TAG)"

build: version
	go build -v bitbucket.org/nirmata/go-vault

dockerBuild: build
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -v -a -installsuffix cgo -ldflags '-w -s' -o $(IMAGE)
	docker build -t $(PREFIX):$(TAG) .

dockerPush: dockerBuild
	docker push $(PREFIX):$(TAG)

clean: 
	go clean -v .

.PHONY: build