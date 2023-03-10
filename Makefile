.PHONY: build
build: ## Builds the binary
	go build -o bin/pyroscope-ci .

.PHONY: test
test: ## Runs the test suite
	go test -race $(shell go list ./...)

.PHONY: test-e2e
test-e2e: ## Runs the e2e test suite
	go test e2e_*.go --tags=e2e

.PHONY: test-install-script
test-install-script: ## Tests the install in a docker container
	docker run -v $(shell PWD)/install.sh:/data/install.sh:ro -it alpine/curl:latest sh -c '/data/install.sh'

.PHONY: godoc
godoc: ## Generate godoc
	godoc -http :8090

.PHONY: lint
lint: ## Run the lint across the codebase
	go run "$(shell scripts/pinned-tools.sh github.com/mgechev/revive)" -config revive.toml -formatter stylish ./...
	staticcheck -f stylish ./...

.PHONY: install-dev-tools
install-dev-tools: ## Install dev tools
	cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI {} go install {}

.PHONY: shellcheck
shellcheck: ## runs shellcheck against install.sh
	shellcheck install.sh


.PHONY: help
help: ## Show this help
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | sed 's/Makefile://' | awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-z0-9A-Z_-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 }'


