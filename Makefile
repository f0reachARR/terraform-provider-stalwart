default: install

# Build the provider
build:
	go build -o terraform-provider-stalwart

# Install the provider locally for testing
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/f0reachARR/stalwart/0.1.0/linux_amd64
	cp terraform-provider-stalwart ~/.terraform.d/plugins/registry.terraform.io/f0reachARR/stalwart/0.1.0/linux_amd64/

# Generate code from OpenAPI spec
generate:
	~/go/bin/oapi-codegen -config oapi-codegen.yaml openapi.yml

# Run tests
test:
	go test -v ./...

# Run acceptance tests
testacc:
	TF_ACC=1 go test -v ./... -timeout 120m

# Clean build artifacts
clean:
	rm -f terraform-provider-stalwart
	rm -rf dist/

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

.PHONY: build install generate test testacc clean fmt lint
