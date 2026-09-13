# go-k8s-layout

An opinionated starting point for Go services that run in Kubernetes.

The repository is two things at once:

- `pkg/` — small, domain-agnostic building blocks: application lifecycle, env-based configuration, structured logging, a pgx pool with a transactor, an HTTP server with health probes.
- `cmd/`, `internal/`, `deploy/` — a working example service (`api`) plus a migrator, built and deployed with the same machinery you would use in production.

Everything here has been exercised end to end against a real cluster: build → image → chart → minikube, including migrations, rolling updates and graceful shutdown.

## Requirements

| Tool | Notes |
|---|---|
| Go 1.26+ | matches the `go` directive in `go.mod` |
| Docker | image builds, local infrastructure |
| kubectl, helm 3+, minikube | Kubernetes workflow (developed against helm 4 and Kubernetes 1.37) |

Project-local tools are installed into `./.bin` (git-ignored):

```
make tools
```

This pins and installs `golangci-lint`, `helmfile` and the `helm-diff` plugin. See `scripts/tools.sh` for versions.

## Repository layout

```
cmd/
  api/                 service entry point (6 lines: background.Run does the rest)
  migrator/            one-shot migration job
internal/
  domain/              entities and domain errors
  repository/postgres/ data access, depends on domain only
  transport/http/      HTTP handlers, declare the interfaces they consume
  app/api/             composition root: config struct and wiring
pkg/
  background/          application lifecycle, signals, graceful shutdown
  env/                 struct-tag configuration parser
  log/                 zerolog wrapper
  database/postgres/   pgx pool, Querier, Transactor
  transport/http/      server, router, middleware, health handler
migrations/
  migrations.go        embed.FS (embed cannot reference parent directories)
  postgres/            goose migrations
deploy/
  docker/              Dockerfile (one, parameterized) and local compose
  k8s/
    charts/            postgres and api charts, shared by every environment
    envs/minikube/     what makes this environment different: helmfile + values
```

The rule that keeps this maintainable: **charts are shared, environments are values**. A new environment is a new directory under `envs/`, never a copy of a chart.

## Quick start: local

Start PostgreSQL, apply migrations, then run the service:

```
make dev-infrastructure-run
```

```
export DB_URL=postgres://test:test@localhost:5432/testdb
export LOG_LEVEL=debug
```

```
go run ./cmd/migrator
```

```
go run ./cmd/api
```

Configuration comes from the environment only — nothing reads a `.env` file at runtime. `cmd/api/.env` exists for the VS Code launch configuration and is git-ignored; exporting the variables above is the shell equivalent.

```
curl -X POST localhost:8080/users -d '{"name":"alice"}'
curl localhost:8080/users
curl localhost:8081/readyz
```

## Quick start: minikube

```
minikube start --cpus=4 --memory=6g --kubernetes-version=v1.37.0
```

```
make k8s-images
```

```
make k8s-up
```

That is the whole sequence. `k8s-images` builds `api` and `migrator` and loads them into the node (there is no registry in this setup); `k8s-up` runs `helmfile apply`, which installs PostgreSQL first, then the API, whose pre-install hook applies migrations before any pod starts.

Reach the service:

```
kubectl port-forward -n dev svc/api 8080:80
```

Reach the database with the same credentials the cluster uses (`test` / `testdb` / `test`):

```
kubectl port-forward -n dev svc/postgres 5433:5432
```

Probes are served on a separate port that is deliberately not exposed through the Service:

```
kubectl port-forward -n dev deployment/api 8081:8081
```

## Make targets

| Target | Purpose |
|---|---|
| `make tools` | install golangci-lint, helmfile, helm-diff into `./.bin` |
| `make lint` / `make test` | Go linting and tests (`make test test_flags="-race"`) |
| `make dev-infrastructure-run` / `-down` | local PostgreSQL via docker compose |
| `make docker-build service=api tag=dev` | build one service image (default tag: short git SHA) |
| `make k8s-images` | build and load both images into minikube |
| `make k8s-diff` | show what would change in the cluster |
| `make k8s-up` / `make k8s-down` | apply or destroy the environment |
| `make k8s-template` | render all manifests to stdout |

Environment switches: `env=minikube` selects `deploy/k8s/envs/<env>`, `image_tag=dev` selects the image tag.

## Configuration

Configuration comes from environment variables only. `pkg/env` maps them onto a struct, joining nested prefixes with `_` and taking defaults from the tag:

```go
type Config struct {
    Postgres postgres.PoolConfig `env:"DB"`    // DB_URL, DB_MAX_CONNS, ...
    HTTP     http.ServerConfig   `env:"HTTP"`  // HTTP_ADDR, HTTP_SHUTDOWN_TIMEOUT
}
```

Main variables:

| Variable | Default | Meaning |
|---|---|---|
| `LOG_LEVEL` | `info` | `debug`, `info`, `warning`, `error` |
| `PRE_SHUTDOWN_DELAY` | `0s` | keep serving after SIGTERM while readiness reports 503 |
| `DB_URL` | — | PostgreSQL DSN, required |
| `DB_MAX_CONNS`, `DB_MIN_CONNS`, `DB_MIN_IDLE_CONNS` | `10`, `5`, `5` | pool sizing |
| `HTTP_ADDR` | `:8080` | application listener |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | graceful shutdown budget for in-flight requests |
| `K8S_HTTP_SERVER_ADDR` | `:8081` | probe listener, separate from the application port |
| `K8S_HTTP_LIVEZ_PATH`, `K8S_HTTP_READYZ_PATH`, `K8S_HTTP_STARTUPZ_PATH` | `/livez`, `/readyz`, `/startupz` | probe paths |
| `K8S_HTTP_READYZ_TIMEOUT` | `1s` | budget for all readiness checks, keep below the probe's `timeoutSeconds` |
| `K8S_HTTP_READYZ_LOG_INTERVAL` | `0s` | reminder interval while a check keeps failing (`0` disables) |
| `MIGRATION_TIMEOUT` | `5m` | migrator only |

In Kubernetes the non-secret variables come from a ConfigMap rendered by the chart, and `DB_URL` is read from a Secret with an explicit `secretKeyRef` — the application never receives the other keys of that Secret.

## Health probes and shutdown

Probes are served on their own port so that application middleware, rate limits or a saturated handler chain cannot affect them:

- `/livez` — no external dependencies. It answers "the process is not wedged". It must never check the database: a database outage would restart every replica and make things worse.
- `/readyz` — checks that make *this instance* unable to serve: a pool ping, plus the application lifecycle flag. Failures are logged on state transitions only (`health check failed` / `health check recovered`), with an optional reminder every `LOG_INTERVAL`.
- `/startupz` — gates liveness and readiness while the process starts.

Shutdown is a sequence, not an event:

```
SIGTERM → readyz starts returning 503
        → PRE_SHUTDOWN_DELAY (still serving traffic)
        → HTTP servers stop accepting, in-flight requests finish
        → dependencies close (LIFO)
```

The delay exists because pod deletion and endpoint removal are not synchronous: kube-proxy on other nodes may still route to this pod for a while after SIGTERM. The chart computes `terminationGracePeriodSeconds` as `preShutdownDelay + shutdownTimeout + buffer` and refuses to render if you override it with a smaller value.

## Migrations

`cmd/migrator` embeds the SQL files and applies them with [goose](https://github.com/pressly/goose), guarded by a PostgreSQL advisory lock so that a manual run cannot race a deployment.

```
go run ./cmd/migrator                    # up (default)
go run ./cmd/migrator -command=status
go run ./cmd/migrator -command=version
go run ./cmd/migrator -command=down
```

New migration: add `migrations/postgres/0000N_name.sql` with `-- +goose Up` and `-- +goose Down` sections.

In Kubernetes the migrator runs as a Helm hook (`pre-install,pre-upgrade`) from an image built out of the same source tree, so the schema and the code that needs it are always released together. `down` is intentionally not part of the hook: roll forward instead.

## Images

One parameterized Dockerfile builds every service:

```
docker build -f deploy/docker/Dockerfile --build-arg SERVICE=api -t api:dev .
```

Multi-stage, `CGO_ENABLED=0`, BuildKit cache mounts for modules and the compiler, and a `distroless/static:nonroot` runtime — around 20 MB, no shell, non-root by default, compatible with a read-only root filesystem and `drop: [ALL]`. Use `kubectl debug` when you need a shell next to the container.

Tags should be immutable (`git rev-parse --short HEAD` is the default in `make docker-build`); `latest` prevents Kubernetes from noticing a new revision.

## Kubernetes layout

```
deploy/k8s/charts/postgres    single-instance StatefulSet for development only
deploy/k8s/charts/api         Deployment, Service, ConfigMap, migration hook, optional HPA/PDB
deploy/k8s/envs/minikube      helmfile.yaml.gotmpl + per-release values
```

The two charts are connected by one contract: the PostgreSQL chart publishes a Secret containing a ready-to-use `DB_URL`, and the API chart consumes it by name. Swap in a managed database or a CloudNativePG cluster and only that Secret's origin changes — the API chart stays untouched.

Each environment pins its own `kubeContext`, so `make k8s-up` cannot apply to the wrong cluster even if `kubectl` points somewhere else.

Both charts ship a `values.schema.json`, so typos and wrong types fail at render time rather than in the cluster.

For a production environment, add `deploy/k8s/envs/prod/` with its own values: images from a registry, `postgres` release disabled in favour of a managed database, secrets from SOPS or External Secrets, HPA and PDB enabled, Ingress instead of port-forward. The charts and the make targets stay the same.

## Things that will bite you

- **Images must be loaded before releases.** Without a registry, `make k8s-images` is what makes `api:dev` exist in the node; otherwise pods end up in `ImagePullBackOff`.
- **PostgreSQL before the API.** The API's migration hook needs the Secret created by the database chart *and* a reachable database. `helmfile` encodes the order with `needs`, but if you run `helm` by hand, install PostgreSQL with `--wait` first.
- **Helm hooks run before the chart's own resources.** A hook Job cannot reference a ServiceAccount or ConfigMap created by the same chart — it will fail with `not found` and Helm will wait until the timeout. Keep hook pods self-contained.
- **`initdb` only runs on an empty volume.** Changing `auth.username` or `auth.database` for an existing release has no effect until the PVC is deleted and recreated.
- **`helm uninstall` does not delete PVCs** created from `volumeClaimTemplates`. `make k8s-down` leaves the data behind on purpose; remove `pvc/data-postgres-0` (or the namespace) for a truly clean start.
- **`kubectl port-forward` is bound to one pod.** It breaks during a rollout — that is the tunnel dying, not your service. Measure availability from inside the cluster.
- **Never remove the minikube container by hand.** `docker rm minikube` or a broad `docker system prune` leaves the profile believing the cluster exists; the recreated node skips `kubeadm init` and kubelet crash-loops. Use `minikube delete`.

## Not included yet

Metrics and tracing, CI, manifest validation with `kubeconform`, secret management, Ingress, and a second (worker) service. The layout has a place for each of them.
