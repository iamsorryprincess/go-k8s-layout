# Dev commands

.PHONY: tools
tools:
	sh scripts/tools.sh

.PHONY: lint
lint:
	./.bin/golangci-lint run ./cmd/... ./internal/... ./pkg/... -c .golangci.yaml
