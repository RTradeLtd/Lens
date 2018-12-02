#LENSVERSION=`git describe --tags`
LENSVERSION="testing"

GOFLAGS=
DIST=$(shell uname)
ifeq ($(DIST), Linux) 
GOFLAGS=-tags gcc7
endif

lens:
	@make cli

.PHONY: deps
deps: 
	@echo "=================== generating dependencies ==================="
	# Install tensorflow
	bash setup/scripts/tensorflow_install.sh

	# Install tesseract
	bash setup/scripts/tesseract_install.sh

	# Update standard dependencies
	dep ensure -v

	# install gofitz
	go get -u $(GOFLAGS) github.com/gen2brain/go-fitz

	# Install counterfeiter, used for mock generation
	go get -u github.com/maxbrunsfeld/counterfeiter
	@echo "===================          done           ==================="

# Build lens cli
.PHONY: cli
cli:
	@echo "====================  building Lens CLI  ======================"
	rm -f temporal-lens
	go build $(GOFLAGS) -ldflags "-X main.Version=$(LENSVERSION)" ./cmd/temporal-lens
	@echo "===================          done           ==================="

# protoc -I protobuf service.proto --go_out=plugins=grpc:protobuf
.PHONY: proto
proto:
	@echo "===================  building protobuffs  ==================="
	# build the request protobuf
	protoc -I=protobuf --go_out=. "protobuf/request.proto"
	protoc -I=protobuf --go_out=. "protobuf/response.proto"
	protoc -I=protobuf --go_out=plugins=grpc:. "protobuf/service.proto"
	@echo "===================          done           ==================="

# Set up test environment
.PHONY: testenv
WAIT=3
testenv:
	@echo "===================   preparing test env    ==================="
	docker-compose -f test/docker-compose.yml up -d
	sleep $(WAIT)
	@echo "===================          done           ==================="

# Run simple checks
.PHONY: check
check:
	go vet $(GOFLAGS) ./...
	go test $(GOFLAGS) -run xxxx ./...

# Generate mocks
.PHONY: mocks
mocks:
	counterfeiter -o ./mocks/manager.mock.go \
		./vendor/github.com/RTradeLtd/rtfs/rtfs.i.go Manager
	counterfeiter -o ./mocks/images.mock.go \
		./analyzer/images/tensorflow.go TensorflowAnalyzer
