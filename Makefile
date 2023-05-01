.PHONY: tidy fmt lint help
.DEFAULT_GOAL:=help

GOLANG_LINT_VERSION=v1.51.2

tidy: ## tidy up go.mod
	go mod tidy

fmt: tidy ## gofmt and goimports all go files
	go run mvdan.cc/gofumpt -l -w -extra .
	find . -name '*.go' -exec go run golang.org/x/tools/cmd/goimports -w {} +

lint: fmt ## run golangcli-lint checks
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_LINT_VERSION) run --timeout 5m

help:  ## Shows help
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'