.PHONY: \
	vet \
	fmt

PKG := github.com/dmitrijs2005/metric-alerting-service

BUILD_VERSION  ?= $(shell git describe --tags --always)
BUILD_DATE     ?= $(shell date +%Y/%m/%d\ %H:%M:%S)
BUILD_COMMIT   ?= $(shell git rev-parse --short HEAD)

LDFLAGS = -X '${PKG}/internal/buildinfo.buildVersion=${BUILD_VERSION}' \
          -X '${PKG}/internal/buildinfo.buildDate=${BUILD_DATE}' \
          -X '${PKG}/internal/buildinfo.buildCommit=${BUILD_COMMIT}'


vet:
	go vet ./...

test:
	go test -v ./...

fmt:
	go fmt ./...

build_server:
	go build -ldflags "$(LDFLAGS)" -o cmd/server/server ${PKG}/cmd/server

build_agent:
	go build -ldflags "$(LDFLAGS)" -o cmd/agent/agent ${PKG}/cmd/agent
