.PHONY: all clean lint test testcov security vuln license addlicense build sbom sign

# Variables
BINARY_NAME=chapa
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=./build
MAIN_PKG=main.go

# TOOLs
GOSEC=$(shell go env GOPATH)/bin/gosec
GOLANGCI_LINT=$(shell go env GOPATH)/bin/golangci-lint
STATICCHECK=$(shell go env GOPATH)/bin/staticcheck
CYCLONEDX=$(shell go env GOPATH)/bin/cyclonedx-gomod
COSIGN=$(shell go env GOPATH)/bin/cosign
SYFT=$(shell go env GOPATH)/bin/syft
ADDLICENSE=$(shell go env GOPATH)/bin/addlicense
GOVULNCHECK=$(shell go env GOPATH)/bin/govulncheck

all:  lint test security vuln license build #sbom sign

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR) $(BUILD_DIR) $(BUILD_DIR)

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-X 'main.Version=${VERSION}'" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)
	cd .dagger && go build  -o ../$(BUILD_DIR)/infra

test:
	@echo "Running tests"
	go test ./...

testcov:
	@echo "Running tests + coverage..."
	@mkdir -p $(BUILD_DIR)
	go test -race -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated - $(BUILD_DIR)/coverage.html"

lint:
	@echo "Linting..."
	$(GOLANGCI_LINT) run --timeout 5m --enable nlreturn --enable nlreturn --enable testifylint --enable revive --enable gci --enable dupl --enable errorlint --enable usestdlibvars --enable wsl --enable perfsprint --enable prealloc --enable makezero --fix  ./...
	 # $(STATICCHECK) ./...

security:
	@echo "Security scan..."
	$(GOSEC) -no-fail -exclude-generated --exclude-dir .dagger  ./...

vuln:
	@echo "Vulnerability scan..."
	$(GOVULNCHECK) ./...

license:
	@echo "Checking source boilerplate..."
	$(ADDLICENSE) --check --ignore 'build/**' --ignore '.dagger/internal/**'  --ignore 'web/**' -c 'The ChapaUY Authors' -l apache  -s=only .

addlicense:
	$(ADDLICENSE)  --ignore 'build/**' --ignore '.dagger/internal/**' --ignore 'web/**'   -c 'The ChapaUY Authors' -l apache  -s=only .

sbom:
	@echo "Generating SBOM..."
	@mkdir -p $(BUILD_DIR)
	$(CYCLONEDX) mod -json -output $(BUILD_DIR)/sbom-cyclonedx.json
	$(SYFT) packages dir:. -o spdx-json=$(BUILD_DIR)/sbom-spdx.json
	@echo "SBOMs generados en directorio $(BUILD_DIR)"

sign:
	@echo "Signing artifact..."
	@if [ ! -f cosign.key ]; then \
		$(COSIGN) generate-key-pair; \
	fi
	$(COSIGN) sign-blob --key cosign.key $(BUILD_DIR)/$(BINARY_NAME) > $(BUILD_DIR)/$(BINARY_NAME).sig
	$(COSIGN) sign-blob --key cosign.key $(BUILD_DIR)/sbom-cyclonedx.json > $(BUILD_DIR)/sbom-cyclonedx.json.sig
	@echo "Artefactos firmados correctamente"

verify:
	$(COSIGN) verify-blob --key cosign.pub $(BUILD_DIR)/$(BINARY_NAME) --signature $(BUILD_DIR)/$(BINARY_NAME).sig
	$(COSIGN) verify-blob --key cosign.pub $(BUILD_DIR)/sbom-cyclonedx.json --signature $(BUILD_DIR)/sbom-cyclonedx.json.sig

INSTALL = go install -v
deps:
	@echo "Installing development dependencies..."
	$(INSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(INSTALL) github.com/securego/gosec/v2/cmd/gosec@latest
	$(INSTALL) golang.org/x/vuln/cmd/govulncheck@latest
	$(INSTALL) honnef.co/go/tools/cmd/staticcheck@latest
	$(INSTALL) github.com/google/addlicense@latest
	$(INSTALL) sigs.k8s.io/bom/cmd/bom@latest
	#$(INSTALL) github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
	#$(INSTALL) github.com/sigstore/cosign/cmd/cosign@latest
	#curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

update:
	@echo "Updating all dependencies..."
	go get -u  && go mod tidy
	cd .dagger && go get -u ./... && go mod tidy
	cd web && pnpm up --latest

#########################
# Frontend (Web)        #
#########################

web-install:
	cd web && pnpm install

web-dev:
	cd web && pnpm run dev

web-build:
	cd web && pnpm run build

web-test:
	cd web && pnpm test
