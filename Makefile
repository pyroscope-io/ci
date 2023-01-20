.PHONY: build
build: ## Builds the binary
	go build -o bin/ci .

.PHONY: test
test: ## Runs the test suite
	go test -race $(shell go list ./...)

.PHONY: godoc
godoc: ## Generate godoc
	godoc -http :8090

.PHONY: help
help: ## Show this help
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | sed 's/Makefile://' | awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-z0-9A-Z_-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 }'


