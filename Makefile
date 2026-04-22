.PHONY: fmt test build docs

fmt:
	gofmt -w main.go internal

test:
	go test ./...

build:
	go build -o terraform-provider-openrouter

docs:
	@echo "Docs are maintained in Terraform Registry markdown layout under ./docs."
