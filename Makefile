LENSVERSION=`git describe --tags`
EDITION=cpu
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
	GO111MODULE=on go mod vendor

	# install gofitz with flags
	GO111MODULE=on go get $(GOFLAGS) github.com/gen2brain/go-fitz
	@echo "===================          done           ==================="

# Build lens cli
.PHONY: cli
cli:
	@echo "====================  building Lens CLI  ======================"
	rm -f temporal-lens
	go build $(GOFLAGS) \
		-ldflags "-X main.Version=$(LENSVERSION) -X main.Edition=$(EDITION)" \
		./cmd/temporal-lens
	@echo "===================          done           ==================="

# Set up test environment
.PHONY: testenv
WAIT=3
testenv:
	@echo "===================   preparing test env    ==================="
	docker-compose -f test/docker-compose.yml up -d
	sleep $(WAIT)
	@echo "===================          done           ==================="

.PHONY: testenv-integration
testenv-integration: testenv
	@echo Connecting testenv IPFS node to RTrade IPFS node for test assets
	ipfs --api=/ip4/127.0.0.1/tcp/5001 swarm connect /ip4/172.218.49.115/tcp/5002/ipfs/Qmf964tiE9JaxqntDsSBGasD4aaofPQtfYZyMSJJkRrVTQ

# Run simple checks
.PHONY: check
check:
	go vet $(GOFLAGS) ./...
	go test $(GOFLAGS) -run xxxx ./...

# Generate code
.PHONY: gen
gen:
	GO111MODULE=on go generate ./...

# Build docker release
.PHONY: docker
docker:
	@echo "===================  building docker image  ==================="
	@echo EDITION: $(EDITION)
	@docker build \
		--build-arg LENSVERSION=$(LENSVERSION)-$(EDITION) \
		--build-arg TENSORFLOW_DIST=$(EDITION) \
		-t rtradetech/lens:$(LENSVERSION)-$(EDITION) .
	@echo "===================          done           ==================="

.PHONY: v2
v2: cli
	./temporal-lens --dev --cfg test/config.json v2
