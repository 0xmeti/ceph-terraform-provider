default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Build provider
.PHONY: build
build:
	go build -o terraform-provider-ceph

# Install provider locally
.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/mahdiinejatii/ceph/1.0.0/linux_amd64
	mv terraform-provider-ceph ~/.terraform.d/plugins/registry.terraform.io/mahdiinejatii/ceph/1.0.0/linux_amd64/

# Generate documentation
.PHONY: docs
docs:
	go generate ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	gofmt -s -w -e .
	terraform fmt -recursive ./examples/

# Run tests
.PHONY: test
test:
	go test -v -cover -timeout=120s -parallel=4 ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf dist/
	rm -f terraform-provider-ceph
