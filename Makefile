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

# Run go tooling commands against code
go-%:
	go $* ./...

.PHONY: lint
lint:
	./hack/install_golint.sh ${BIN_LOCATION} ${GO_LINTER_VERSION}
	${BIN_LOCATION}/golangci-lint run ./... --fix

.PHONY: binaries
binaries:
	./hack/install_goreleaser.sh ${BIN_LOCATION} ${GO_RELEASER_VERSION}
	${BIN_LOCATION}/goreleaser build --snapshot --clean


.PHONY: coco
coco:
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