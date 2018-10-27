.DEFAULT_GOAL: build

GIT_VERSION := $(shell git describe --dirty --always --tags)
TAG?=$(GIT_VERSION)
IMAGE?=kube-vault-client
PREFIX?=nirmata/$(IMAGE)
ARCH?=amd64

build:
	@echo "version: $(PREFIX):$(TAG)"
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -a -installsuffix cgo -ldflags '-w -s' -o $(IMAGE)
	docker build -t $(PREFIX):$(TAG) .

dockerBuild: build

dockerPush: dockerBuild
	docker push $(PREFIX):$(TAG)

dockerTagLatest: dockerPush
	docker tag  $(PREFIX):$(TAG)  $(PREFIX):latest
	docker push $(PREFIX):latest

clean:
	go clean -v .
	docker rmi $(PREFIX):$(TAG)

.PHONY: build