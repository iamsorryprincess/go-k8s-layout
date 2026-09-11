# Dev commands

.PHONY: tools
tools:
	sh scripts/tools.sh

.PHONY: lint
lint:
	./.bin/golangci-lint run ./cmd/... ./internal/... ./pkg/... -c .golangci.yaml

# Extra flags for the test target, e.g. make test test_flags="-race -v"
test_flags =

.PHONY: test
test:
	go test -count=1 $(test_flags) ./internal/... ./pkg/...

# Backend dev deploy dir
backend_dev_dir = deploy/docker

.PHONY: dev-infrastructure-run
dev-infrastructure-run:
	docker compose -f $(backend_dev_dir)/docker-compose.yaml -p dev-infrastructure up -d

.PHONY: dev-infrastructure-down
dev-infrastructure-down:
	docker compose -f $(backend_dev_dir)/docker-compose.yaml -p dev-infrastructure down
