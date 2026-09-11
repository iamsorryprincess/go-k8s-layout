# Dev commands

.PHONY: tools
tools:
	sh scripts/tools.sh

.PHONY: lint
lint:
	./.bin/golangci-lint run ./cmd/... ./internal/... ./pkg/... -c .golangci.yaml

# Backend dev deploy dir
backend_dev_dir = deploy/dev

.PHONY: dev-infrastructure-run
dev-infrastructure-run:
	docker compose -f $(backend_dev_dir)/docker-compose.yml -p dev-infrastructure up -d

.PHONY: dev-infrastructure-down
dev-infrastructure-down:
	docker compose -f $(backend_dev_dir)/docker-compose.yml -p dev-infrastructure down
