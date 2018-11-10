#LENSVERSION=`git describe --tags`
LENSVERSION="testing"

lens:
	@make cli

# Build lens cli
.PHONY: cli
cli:
	@echo "===================  building Lens CLI  ==================="
	rm -f temporal-lens
	go build -ldflags "-X main.Version=$(LENSVERSION)" ./cmd/temporal-lens
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
	docker-compose -f lens.yml up -d
	sleep $(WAIT)
	@echo "===================          done           ==================="
