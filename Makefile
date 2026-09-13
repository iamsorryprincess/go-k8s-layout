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

# Docker image build, e.g. make docker-build service=api tag=dev
docker_dir = deploy/docker
service = api
tag = $(shell git rev-parse --short HEAD)

.PHONY: docker-build
docker-build:
	docker build -f $(docker_dir)/Dockerfile --build-arg SERVICE=$(service) -t $(service):$(tag) .

# Puts the image into minikube so the cluster can run it without a registry
.PHONY: minikube-load
minikube-load: docker-build
	minikube image load $(service):$(tag)

# Kubernetes environment, e.g. make k8s-up env=minikube image_tag=dev
env = minikube
image_tag = dev
helmfile = ./.bin/helmfile -f deploy/k8s/envs/$(env)/helmfile.yaml.gotmpl -e $(env)

.PHONY: k8s-images
k8s-images:
	$(MAKE) minikube-load service=api tag=$(image_tag)
	$(MAKE) minikube-load service=migrator tag=$(image_tag)

.PHONY: k8s-diff
k8s-diff:
	IMAGE_TAG=$(image_tag) $(helmfile) diff

.PHONY: k8s-up
k8s-up:
	IMAGE_TAG=$(image_tag) $(helmfile) apply

.PHONY: k8s-down
k8s-down:
	IMAGE_TAG=$(image_tag) $(helmfile) destroy

.PHONY: k8s-template
k8s-template:
	IMAGE_TAG=$(image_tag) $(helmfile) template
