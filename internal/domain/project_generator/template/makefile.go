package template

var TplMakefile = `GO=go
GOMODULE=%%GO-MODULE%%
debug?=0 # 0: release, 1: support dlv debug

# Check if you have docker permissions
DOCKER := docker
DOCKER_PERMISSION := $(shell [ ! -w /var/run/docker.sock ] && echo "NO")
ifeq ($(DOCKER_PERMISSION), NO)
	DOCKER = sudo docker
endif
IMAGE_NAME=<your-registry>/%%PROJECT-NAME%%

# From gitlab CI
CI_COMMIT_REF_NAME ?= ""
CI_COMMIT_SHORT_SHA ?= ""
PIPE_DATE ?= $(shell date +%Y%m%d)

GIT_TAG    = $(if $(filter-out "",${CI_COMMIT_REF_NAME}),${CI_COMMIT_REF_NAME},$(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null || echo ""))
GIT_BRANCH = $(shell git symbolic-ref --short HEAD || echo "")
GIT_HASH  = $(if $(filter-out "",${CI_COMMIT_SHORT_SHA}),${CI_COMMIT_SHORT_SHA},$(shell git rev-parse --short=8 HEAD || echo ""))

# if GIT_TAG is empty, use git branch and git commit hash as version
VERSION = $(if $(filter-out "",${GIT_TAG}),${GIT_TAG}-${PIPE_DATE}-${GIT_HASH},${GIT_BRANCH}-${PIPE_DATE}-${GIT_HASH})

LDFLAGS += -X "${GOMODULE}/pkg/version.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "${GOMODULE}/pkg/version.GitHash=${GIT_HASH}"
LDFLAGS += -X "${GOMODULE}/pkg/version.GitBranch=${GIT_BRANCH}"
LDFLAGS += -X "${GOMODULE}/pkg/version.GitTag=${GIT_TAG}"
LDFLAGS += -X "${GOMODULE}/pkg/version.GitDirty=$(shell test -n "` + "`" + `git status --porcelain` + "`" + `" && echo "true")"
LDFLAGS += -X "${GOMODULE}/pkg/version.Debug=$(shell if [ ${debug} -eq 1 ]; then echo "true"; fi)"
LDFLAGS += -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn

ifeq ($(debug), 1)
	GC_FLAGS:=-gcflags "all=-N -l"
else
	GO_LDFLAGS += -s -w
	GO_TRIMPATH = -trimpath
endif


TARGETS := $(shell ls cmd)

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/swaggo/swag/cmd/swag@latest
	go get github.com/google/wire/cmd/wire@latest
	go get github.com/golang/mock/mockgen@v1.6.0
	go get github.com/swaggo/swag/cmd/swag@latest

.PHONY: swag
# generate api doc
swag:
ifeq ($(wildcard internal/handler/http),internal/handler/http) 
	swag fmt -dir ./internal/handler/http/
	swag init -g ./internal/server/http/router.go -o ./docs --parseDependency -q --exclude ./_manifest
endif

.PHONY: pb
# generate api proto
pb:
	rm -rf ./api/%%PROJECT-NAME-DIR%%
	protoc \
	  --proto_path ./proto \
	  --proto_path ./third_party/proto \
	  --go_out=./api \
	  --go-grpc_out=./api \
	  --validate_out=lang=go:./api \
	  ./proto/*.proto

generate: pb swag
# generate
generate:
	go get -u gorm.io/gorm
	go get -u gorm.io/driver/postgres
	go get -u gorm.io/driver/mysql
	GOFLAGS=-mod=mod GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn go generate ./...
	go mod tidy

.PHONY: vet
# vet code
vet:
	go vet ./...

build: vet
# build binary
build: $(TARGETS)

$(TARGETS):
	${GO} build -ldflags '${GO_LDFLAGS} ${LDFLAGS}' ${GC_FLAGS} ${GO_TRIMPATH} -o build/bin/ ${GOMODULE}/cmd/$@

.PHONY: clear
# clear binary
clear:
	rm -f build/bin/*

.PHONY: run
# run binary
run:
	./build/bin/%%PROJECT-NAME%% -config ./configs/%%PROJECT-NAME%%-local.yml

.PHONY: test
# go test
test:
	go test $(shell go list ./... | grep internal)	

.PHONY: image
# build docker image
image:
	go mod vendor
	docker build \
	  --build-arg ci_commit_ref_name=${CI_COMMIT_REF_NAME} \
	  --build-arg ci_commit_short_sha=${CI_COMMIT_SHORT_SHA} \
	  --build-arg pipe_date=${PIPE_DATE} \
	  --no-cache --network=host \
	  -f build/package/Dockerfile \
	  -t ${IMAGE_NAME}:${VERSION} .
	rm -rf vendor

.PHONY: push
# push docker image
push:
	docker push ${IMAGE_NAME}:${VERSION}

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
`
