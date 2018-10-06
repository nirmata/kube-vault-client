.DEFAULT_GOAL: build

TAG=latest
PREFIX?=registry-v2.nirmata.io/nirmata/nirmata-vault
ARCH?=amd64

build:
	go build -v bitbucket.org/nirmata/go-vault


docker: 
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -v -a -installsuffix cgo -ldflags '-w -s' -o vault-client
	docker build -t $(PREFIX):$(TAG) .
	docker push $(PREFIX):$(TAG)

clean: 
	go clean -v .

.PHONY: build