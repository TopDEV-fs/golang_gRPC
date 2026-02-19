# Product Catalog Service

Test task implementation for a DDD/Clean Architecture Go service using Spanner, gRPC, CQRS, CommitPlan, and transactional outbox.

## Stack

- Go 1.21+
- Google Cloud Spanner (emulator for local)
- gRPC + Protocol Buffers
- `github.com/Vektor-AI/commitplan` (+ Spanner driver)
- `math/big` for decimal arithmetic (`*big.Rat`)
- `testify` for tests
- `golangci-lint` for static checks
- GitHub Actions CI

## Architecture Highlights

- **Domain purity:** no DB/proto/framework imports in domain.
- **Golden Mutation Pattern:** usecases build plan and apply transaction atomically.
- **CQRS:** commands go through aggregate, queries use read model DTOs.
- **Transactional outbox:** domain events persisted in `outbox_events` in same write transaction.
- **Change tracking:** aggregate marks dirty fields; repository emits targeted updates.

## Project Structure

- [cmd/server/main.go](cmd/server/main.go)
- [internal/app/product/domain/product.go](internal/app/product/domain/product.go)
- [internal/app/product/usecases/create_product/interactor.go](internal/app/product/usecases/create_product/interactor.go)
- [internal/app/product/queries/get_product/query.go](internal/app/product/queries/get_product/query.go)
- [internal/app/product/repo/product_repo.go](internal/app/product/repo/product_repo.go)
- [internal/transport/grpc/product/handler.go](internal/transport/grpc/product/handler.go)
- [proto/product/v1/product_service.proto](proto/product/v1/product_service.proto)
- [migrations/001_initial_schema.sql](migrations/001_initial_schema.sql)
- [tests/e2e/product_test.go](tests/e2e/product_test.go)
- [docs/architecture.md](docs/architecture.md)
- [.github/workflows/ci.yml](.github/workflows/ci.yml)
- [.golangci.yml](.golangci.yml)
- [Dockerfile](Dockerfile)

## Run locally

### 1) Start Spanner emulator

```bash
docker-compose up -d
```

### 2) Create instance/database and apply migration

Use gcloud (emulator configured):

```bash
make migrate
```

### 3) Run tests

```bash
set SPANNER_EMULATOR_HOST=localhost:9010
make test
```

Run only unit tests (excluding e2e):

```bash
go test $(go list ./... | grep -v '/tests/e2e')
```

### 4) Start gRPC server

```bash
set SPANNER_EMULATOR_HOST=localhost:9010
set SPANNER_DATABASE=projects/local-dev/instances/test-instance/databases/product_catalog
make run
```

## Design Decisions / Trade-offs

- Kept handlers thin and focused on validation/mapping only.
- Repositories return mutations, commit is done only in usecases.
- Added manual proto Go stubs in repo so project is self-contained; replace with generated files in production.
- E2E tests target emulator and auto-provision test DBs; they skip when emulator is unavailable.
- Added public facade package [pkg/productcatalog/service.go](pkg/productcatalog/service.go) for clearer composition boundary.

## Quality Gates

- Lint: `golangci-lint run`
- Vet: `go vet ./...`
- CI pipeline: lint + vet + unit tests on PRs/pushes
- Container image build: [Dockerfile](Dockerfile)

## Notes

- Outbox processor/PubSub integration intentionally not implemented per task scope.
- AuthN/AuthZ and observability are intentionally omitted.