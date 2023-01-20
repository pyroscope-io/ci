.PHONY: build
build: ## Builds the binary
	go build -o bin/ci .

.PHONY: test
test: ## Runs the test suite
	go test -race $(shell go list ./...)

.PHONY: godoc
godoc: ## Generate godoc
	godoc -http :8090

.PHONY: lint
lint: ## Run the lint across the codebase
	go run "$(shell scripts/pinned-tools.sh github.com/mgechev/revive)" -config revive.toml -formatter stylish ./...
#	staticcheck -f stylish ./...

.PHONY: install-dev-tools
install-dev-tools: ## Install dev tools
	cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI {} go install {}

.PHONY: help
help: ## Show this help
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | sed 's/Makefile://' | awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-z0-9A-Z_-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 }'


