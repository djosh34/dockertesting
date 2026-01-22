.PHONY: help test test-all

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Run unit tests (excludes integration tests)
	go test -v -tags '!integration' ./...

test-all: ## Run all tests (unit + integration)
	go test -v -tags integration ./...
