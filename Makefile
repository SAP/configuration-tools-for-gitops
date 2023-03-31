# check here for makefile inspiration:
# https://github.com/argoproj/argo-cd/blob/master/Makefile

CURRENT_DIR=$(shell pwd)
BIN_LOCATION=${HOME}/go/bin
PATH:=$(PATH):$(PWD)/hack

GO_LINTER_VERSION=$(shell cat ${CURRENT_DIR}/.buildvars.yml | yq '.golangci_version')
GO_RELEASER_VERSION=$(shell cat ${CURRENT_DIR}/.buildvars.yml | yq '.goreleaser_version')


VERSION_PACKAGE=github.com/configuration-tools-for-gitops/pkg/version

# env variables
GOPATH          ?=$(shell if test -x `which go`; then go env GOPATH; else echo "$(HOME)/go"; fi)
GOCACHE         ?=$(HOME)/.cache/go-build


.PHONY: test
test: ## Run unit tests in the code base outside of the tmp/ folder
	test_dir="$(shell go list ./... | grep -v -e tmp/)"; go test -v -race $$test_dir -coverprofile -covermode=count -coverprofile=coverage.out fmt
	go tool cover -func=coverage.out -o=coverage.out
	cat coverage.out

package?=""
function?=""
.PHONY: test-this
test-this: ## Run unit tests for the package specified as argument (e.g. make test-this package=cmd/coco/dependencies function=TestGraph)
	./hack/test_this.sh ${package} ${function}


.PHONY: integration
integration: ## Run integration tests in the code base outside of the tmp/ folder
	test_dir="$(shell go list ./... | grep -v -e /tmp)"; INTEGRATION_TESTS=true go test --run Integration -timeout 300s -race $$test_dir -coverprofile cover.out
	go tool cover -func ./cover.out 
.PHONY: test-all
test-all: ## Run all tests in the code base outside of the tmp/ folder
	test_dir="$(shell go list ./... | grep -v /tmp)"; INTEGRATION_TESTS=true go test -run .\* -v -timeout 300s -race $$test_dir -coverprofile cover.out
	go tool cover -func ./cover.out 

go-%: ## Run a go toolchain command against the code base (e.g. go fmt ./...)
	go $* ./...

.PHONY: lint
lint: ## Run golangci-lint (configuration in .golangci.yml)
	./hack/install_golint.sh ${BIN_LOCATION} ${GO_LINTER_VERSION}
	${BIN_LOCATION}/golangci-lint run ./... --fix

.PHONY: binaries
binaries: ## Build the coco binaries for all target architectures (result is stored in ./dist/coco_$os_$arch/coco)
	./hack/install_goreleaser.sh ${BIN_LOCATION} ${GO_RELEASER_VERSION}
	${BIN_LOCATION}/goreleaser build --snapshot --clean


.PHONY: coco
coco: ## Build the coco binary for the current architecture (result is stored in ./dist/coco_$os_$arch/coco)
	./hack/install_goreleaser.sh ${BIN_LOCATION} ${GO_RELEASER_VERSION}
	${BIN_LOCATION}/goreleaser build --single-target --snapshot --clean


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

.PHONY: help
help:
	@grep -E '^[%a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
