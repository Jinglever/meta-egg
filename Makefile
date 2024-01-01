GO=go
GOMODULE=meta-egg
debug?=0 # 0: release, 1: support dlv debug

LDFLAGS += -X "${GOMODULE}/pkg/version.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "${GOMODULE}/pkg/version.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "${GOMODULE}/pkg/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
LDFLAGS += -X "${GOMODULE}/pkg/version.GitTag=$(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)"
LDFLAGS += -X "${GOMODULE}/pkg/version.GitDirty=$(shell test -n "`git status --porcelain`" && echo "true")"
LDFLAGS += -X "${GOMODULE}/pkg/version.Debug=$(shell if [ ${debug} -eq 1 ]; then echo "true"; fi)"

LDFLAGS_IE = -X "${GOMODULE}/pkg/version.Release=IE"

ENV_WINDOWS_64 = GOOS=windows GOARCH=amd64
ENV_WINDOWS_ARM64 = GOOS=windows GOARCH=arm64
ENV_LINUX_64 = GOOS=linux GOARCH=amd64
ENV_LINUX_ARM64 = GOOS=linux GOARCH=arm64
ENV_DARWIN_64 = GOOS=darwin GOARCH=amd64
ENV_DARWIN_ARM64 = GOOS=darwin GOARCH=arm64

ifeq ($(OS),Windows_NT)
    ENV_CURRENT = $(ENV_WINDOWS_64)
    ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
        ENV_CURRENT = $(ENV_WINDOWS_64)
    endif
    ifeq ($(PROCESSOR_ARCHITECTURE),ARM64)
        ENV_CURRENT = $(ENV_WINDOWS_ARM64)
    endif
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        ENV_CURRENT = $(ENV_LINUX_64)
        ifeq ($(shell uname -m),arm64)
            ENV_CURRENT = $(ENV_LINUX_ARM64)
        endif
    endif
    ifeq ($(UNAME_S),Darwin)
        ENV_CURRENT = $(ENV_DARWIN_64)
        ifeq ($(shell uname -m),arm64)
            ENV_CURRENT = $(ENV_DARWIN_ARM64)
        endif
    endif
endif

$(info ENV_CURRENT=$(ENV_CURRENT))

ifeq ($(debug), 1)
	GC_FLAGS:=-gcflags "all=-N -l"
else
	GO_LDFLAGS += -s -w
	GO_TRIMPATH = -trimpath
endif

TARGETS := $(shell ls cmd)

.PHONY: vet
# vet code
vet:
	go vet ./...

build: vet
# build binary
build:
	${ENV_CURRENT} ${GO} build -ldflags '${GO_LDFLAGS} ${LDFLAGS} ${LDFLAGS_IE}' ${GC_FLAGS} ${GO_TRIMPATH} -o build/bin/meta-egg ${GOMODULE}/cmd/meta-egg

release: vet
# release binary
release: $(TARGETS)

$(TARGETS):
	${ENV_WINDOWS_64} ${GO} build -ldflags '${GO_LDFLAGS} ${LDFLAGS}' ${GC_FLAGS} ${GO_TRIMPATH} -o build/bin/$@_win_amd64 ${GOMODULE}/cmd/$@
	${ENV_WINDOWS_ARM64} ${GO} build -ldflags '${GO_LDFLAGS} ${LDFLAGS}' ${GC_FLAGS} ${GO_TRIMPATH} -o build/bin/$@_win_arm64 ${GOMODULE}/cmd/$@
	${ENV_LINUX_64} ${GO} build -ldflags '${GO_LDFLAGS} ${LDFLAGS}' ${GC_FLAGS} ${GO_TRIMPATH} -o build/bin/$@_linux_amd64 ${GOMODULE}/cmd/$@
	${ENV_LINUX_ARM64} ${GO} build -ldflags '${GO_LDFLAGS} ${LDFLAGS}' ${GC_FLAGS} ${GO_TRIMPATH} -o build/bin/$@_linux_arm64 ${GOMODULE}/cmd/$@
	${ENV_DARWIN_64} ${GO} build -ldflags '${GO_LDFLAGS} ${LDFLAGS}' ${GC_FLAGS} ${GO_TRIMPATH} -o build/bin/$@_darwin_amd64 ${GOMODULE}/cmd/$@
	${ENV_DARWIN_ARM64} ${GO} build -ldflags '${GO_LDFLAGS} ${LDFLAGS}' ${GC_FLAGS} ${GO_TRIMPATH} -o build/bin/$@_darwin_arm64 ${GOMODULE}/cmd/$@

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