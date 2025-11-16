# Repository Guidelines

## Project Structure & Module Organization
Public protobufs live in `api/` (HTTP, gRPC, and OpenAPI stubs are generated in-place), while runnable entry points are under `cmd/` such as `cmd/user`. Layered business code sits in `internal/`: `biz/` orchestrates workflows, `service/` handles transport adapters, `data/` speaks to databases and RPCs, `bo/` carries transformed payloads, `conf/` holds typed config, and `server/` wires everything together. Shared YAML lives in `configs/`, local Docker assets in `deployments/`, and additional protos in `third_party/`.

## Build, Test, and Development Commands
- `make init` sets up the Lynx CLI plus proto and Wire generators; run once per machine.
- `make api`, `make config`, and `make validate` regenerate protobufs; commit both the `.proto` and generated Go/openapi artifacts.
- `make build` (or `go build ./...`) drops binaries into `bin/` with the current `git describe` version embedded.
- `docker compose -f deployments/docker-compose.local.yml up -d` launches PostgreSQL/Redis for local runs; pair it with `go run ./cmd/user -conf ./configs/bootstrap.local.yaml`.
- `make run` executes the full pipeline (proto generation + `kratos run`) when validating Polaris-backed environments.

## Coding Style & Naming Conventions
All Go files must be `gofmt`-clean; run `go fmt ./...` before committing and prefer `goimports` to maintain stable import blocks. File names stay lowercase with optional underscores (`user_service.go`), exported identifiers use PascalCase, and package-scoped errors follow `errFoo`. Dependency wiring should rely on Google Wire in `internal/server`, and generated files must be edited only via the corresponding Make targets to avoid drift.

## Testing Guidelines
Create table-driven `_test.go` files next to the code under test, focusing on `internal/biz` decision paths and `internal/service` adapters. Use `go test ./...` (optionally `-race`) before each PR; spin up the Docker Compose stack when tests require Postgres or Redis.

## Commit & Pull Request Guidelines
Follow Conventional Commits (`fix:`, `feat:`, `chore(deps):`), keep subjects under 72 characters, and group unrelated changes into separate commits so changelog tooling can slice releases cleanly. A PR should describe intent, link issues, call out config toggles, and attach evidence of the commands you ran (`go test`, `make api`, etc.); include curl traces or screenshots when altering externally visible APIs.

## Environment & Configuration Tips
Use Go 1.25.3 (`go env -w GOTOOLCHAIN=go1.25.3`) to match CI and avoid module incompatibilities. Prefer `configs/bootstrap.local.yaml` for local work, reserve `configs/bootstrap.yaml` for Polaris deployments, and keep credentials out of Git by supplying them via environment variables or secret managers.
