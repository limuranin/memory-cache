.PHONY: test test_integration build dep help

test: ## Run unittests
	@go test -v -count=1 -gcflags=all=-l `go list ./... | grep -v -e tests`

test_integration: ## Start server and run integration test
	@go test -v -p=1 -count=1 ./tests

build: ## Build memory cache to cmd/memory-cache/
	@go build -o ./cmd/memory-cache/memory-cache ./cmd/memory-cache

dep: ## Get dependencies
	@go mod download

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
