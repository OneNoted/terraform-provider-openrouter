# List available tasks.
default:
    @just --list

# Format Go source.
fmt:
    gofmt -w main.go internal

# Run unit tests.
test:
    go test ./...

# Build the provider binary.
build:
    go build -o terraform-provider-openrouter

# Run static checks.
vet:
    go vet ./...

# Run vulnerability scan.
vuln:
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Run the standard local verification suite.
check: fmt test vet build vuln
    rm -f terraform-provider-openrouter

# Docs are maintained in Terraform Registry markdown layout.
docs:
    @echo "Docs are maintained in Terraform Registry markdown layout under ./docs."
