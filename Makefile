#
# Copyright 2023 Comcast Cable Communications Management, LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: Apache-2.0
#
GOARCH ?= amd64
GOOS ?= linux
GOHOSTARCH = $(shell go env GOHOSTARCH)
GOHOSTOS = $(shell go env GOHOSTOS)

BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
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
