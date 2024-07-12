.PHONY: help
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | awk -F ':' '{ printf " %-20s %s\n", $$1, $$2 }'

## build-project: Build the project and produce a binary
.PHONY: build-project
build-project:
	mkdir -p bin
	go build -o ./bin/dockerExec ./cmd/dockerExec/

## go-tests: Runs all golang tests for all packages
.PHONY: go-tests
go-tests:
	go test ./... -json | tparse -all
