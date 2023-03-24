# check here for makefile inspiration:
# https://github.com/argoproj/argo-cd/blob/master/Makefile

CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist

VERSION_PACKAGE=github.com/configuration-tools-for-gitops/pkg/version
HOST_OS=$(shell go env GOOS)
HOST_ARCH=$(shell go env GOARCH)

GO_LINTER_VERSION=v1.51.2

# env variables

VERSION					=$(shell cat ${CURRENT_DIR}/VERSION)

BUILD_DATE      =$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT      ?=$(shell git rev-parse HEAD)
GIT_TAG         ?=$(shell if [ -z "`git status --porcelain`" ]; then echo v/$(VERSION); fi)
GIT_TREE_STATE  ?=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
VOLUME_MOUNT    =$(shell if test "$(go env GOOS)" = "darwin"; then echo ":delegated"; elif test selinuxenabled; then echo ":delegated"; else echo ""; fi)


GOPATH          ?=$(shell if test -x `which go`; then go env GOPATH; else echo "$(HOME)/go"; fi)
GOCACHE         ?=$(HOME)/.cache/go-build

dist/coco-darwin-amd64: GOARGS = GOOS=darwin GOARCH=amd64
dist/coco-linux-amd64: GOARGS = GOOS=linux GOARCH=amd64
dist/coco-darwin-arm64: GOARGS = GOOS=darwin GOARCH=arm64
dist/coco-linux-arm64: GOARGS = GOOS=linux GOARCH=arm64

GIT_USER        ?=$(shell git config user.name)
GIT_USER_LOWER  =$(shell echo $(GIT_USER) | tr '[A-Z]' '[a-z]')

DOCKER_SRCDIR   ?=$(GOPATH)/src
DOCKER_WORKDIR  ?=$(CURRENT_DIR)

PATH:=$(PATH):$(PWD)/hack

# docker image publishing options
DOCKER_PUSH ?= false
# perform static compilation
STATIC_BUILD ?= true
# build development images
DEV_IMAGE ?= false

override LDFLAGS += \
  -X ${VERSION_PACKAGE}.version=${VERSION} \
  -X ${VERSION_PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${VERSION_PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${VERSION_PACKAGE}.gitTreeState=${GIT_TREE_STATE}

ifeq (${STATIC_BUILD}, true)
override LDFLAGS += -extldflags "-static"
endif

ifneq (${GIT_TAG},)
IMAGE_TAG=${VERSION}
LDFLAGS += -X ${VERSION_PACKAGE}.gitTag=${GIT_TAG}
else
IMAGE_TAG?=latest
endif

USER_ID =$(id -u ${USER})
GROUP_ID =$(id -g ${USER})

# Run tests
.PHONY: test
test:
	test_dir="$(shell go list ./... | grep -v -e tmp/)"; go test -v -race $$test_dir -coverprofile cover.out
	go tool cover -func ./cover.out 

# Run integration tests
.PHONY: integration
integration:
	test_dir="$(shell go list ./... | grep -v -e /tmp)"; INTEGRATION_TESTS=true go test --run Integration -timeout 300s -race $$test_dir -coverprofile cover.out
	go tool cover -func ./cover.out 
.PHONY: test-all
test-all:
	test_dir="$(shell go list ./... | grep -v /tmp)"; INTEGRATION_TESTS=true go test -run .\* -v -timeout 300s -race $$test_dir -coverprofile cover.out
	go tool cover -func ./cover.out 

# Run go fmt against code
fmt:
	go fmt -s ./...

# Run go vet against code
vet:
	go vet ./...

# Cleans VSCode debug.test files from sub-dirs to prevent them from being included in packr boxes
.PHONY: clean-debug
clean-debug:
	-find ${CURRENT_DIR} -name debug.test | xargs rm -f

.PHONY: clean
clean: clean-debug
	-rm -rf ${CURRENT_DIR}/dist
	-rm -rf coco
	-rm -rf cover.out
	-rm -rf vet.log


.PHONY: lint
lint:
	@ $(eval install_to := ${HOME}/go/bin)
	./hack/validate_golint_version.sh ${install_to} ${GO_LINTER_VERSION}
	${install_to}/golangci-lint run ./... --fix

dist/coco-%.tar.gz: dist/coco-%
	tar -czvf dist/coco-$*.tar.gz dist/coco-$*
dist/coco-%:
	CGO_ENABLED=0 $(GOARGS) go build -v -gcflags '${GCFLAGS}' -ldflags '${LDFLAGS} -extldflags -static' -o $@ ./cmd/coco

.PHONY: coco
coco:
	CGO_ENABLED=0 go build -v -gcflags '${GCFLAGS}' -ldflags '${LDFLAGS} -extldflags -static' -o ${DIST_DIR}/coco ./cmd/coco
