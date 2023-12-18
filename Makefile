GOARCH ?= amd64
GOOS ?= linux
GOHOSTARCH = $(shell go env GOHOSTARCH)
GOHOSTOS = $(shell go env GOHOSTOS)

BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
# must be "Version", NOT "VERSION" to be consistent with xpc jenkins env
Version ?= $(shell git log -1 --pretty=format:"%h")
BUILDTIME := $(shell date -u +"%F_%T_%Z")

all: build

build:  ## Build a version
	go build -v -ldflags="-X xconfadmin/common.BinaryBranch=${BRANCH} -X xconfadmin/common.BinaryVersion=${Version} -X xconfadmin/common.BinaryBuildTime=${BUILDTIME}" -o bin/xconfadmin-${GOOS}-${GOARCH} main.go

test:
	ulimit -n 10000 ; go test ./... -cover -count=1

cover:
	go test ./... -count=1 -coverprofile=coverage.out

html:
	go tool cover -html=coverage.out

clean: ## Remove temporary files
	go clean
	go clean --testcache

release:
	go build -v -ldflags="-X xconfadmin/common.BinaryBranch=${BRANCH} -X xconfadmin/common.BinaryVersion=${Version} -X xconfadmin/common.BinaryBuildTime=${BUILDTIME}" -o bin/xconfadmin-${GOOS}-${GOARCH} main.go
