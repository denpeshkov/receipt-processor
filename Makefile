.PHONY: all
all: tidy lint test run

.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: tidy
tidy: ## Tidy
	go mod tidy -v
	go mod verify
	go fmt ./... 
	go vet ./...

.PHONY: lint
lint: ## Lint
	@docker run -t --rm -v .:/app -v ~/.cache/golangci-lint/v1.61.0:/root/.cache -w /app golangci/golangci-lint:v1.61.0 golangci-lint run -v -c .golangci.yml

.PHONY: test
test: ## Test
	go test -race ./...

.PHONY: run
run: ## Run service on localhost:8080 in debug mode
	@go run ./cmd/processor -addr=localhost:8080 -debug=true