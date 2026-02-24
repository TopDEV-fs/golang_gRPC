# Architecture Notes

## Layering

```
cmd/server          – process entry point; wires deps, starts gRPC server
internal/
  app/product/
    domain/         – pure aggregate, value objects, domain events, errors
    domain/services – stateless domain services (e.g. PricingCalculator)
    contracts/      – interface definitions for repositories and read models
    outbox/         – shared helper: converts domain events → Spanner mutations
    usecases/       – command interactors (one package per use case)
    queries/        – read-side query handlers (one package per query)
    repo/           – Spanner implementations of contracts interfaces
  models/           – Spanner table/column name constants (no domain logic)
  services/         – composition root: wires all deps into a Container
  transport/grpc/   – gRPC handlers; validation + delegation only
  pkg/              – cross-cutting utilities (clock, committer)
pkg/productcatalog  – public facade over internal/services for embedding
proto/              – protobuf definitions and generated stubs
migrations/         – DDL scripts applied by gcloud spanner
tests/e2e/          – end-to-end tests requiring the Spanner emulator
```

## Golden Mutation Pattern

Every command use case follows a deterministic five-step flow:

1. **Load** – call `repo.FindByID` or construct a new aggregate.
2. **Mutate** – invoke the aggregate business method; invariants are checked here.
3. **Plan** – call `repo.InsertMut` / `repo.UpdateMut` to get Spanner mutations.
4. **Outbox** – call `outbox.BuildMuts` to serialise `PullDomainEvents()` into outbox mutations.
5. **Commit** – call `committer.Apply(ctx, plan)` once; all mutations land in one transaction.

This pattern guarantees that aggregate writes and outbox events are always atomic.

## Transactional Outbox

Domain events are accumulated inside the aggregate in memory (`events []DomainEvent`).  
They are consumed by calling `PullDomainEvents()` (consuming pattern – clears the slice).  
The `outbox.BuildMuts` helper serialises each event to JSON and returns a slice of
`*spanner.Mutation` to be added to the same commit plan as the aggregate mutation.

## CQRS Separation

| Side    | Path                                 | Pattern                          |
|---------|--------------------------------------|----------------------------------|
| Command | `usecases/<name>/interactor.go`      | Aggregate + Golden Mutation      |
| Query   | `queries/<name>/query.go`            | Direct SQL → flat DTO mapping    |

Queries bypass the domain aggregate entirely; they return pre-formatted strings
(e.g. `"19.99"`) resolved at query time with effective-price calculation.

## Change Tracking

The `ChangeTracker` embedded in every aggregate records which scalar fields have been
mutated. `ProductRepo.UpdateMut` reads the tracker to emit a targeted
`spanner.UpdateMap` covering only dirty columns, avoiding unnecessary writes.

## Money and Discount Precision

All monetary amounts use `*big.Rat` to avoid floating-point precision errors.  
Prices are stored as (numerator, denominator) INT64 pairs in Spanner.  
Discount percentages are stored as `NUMERIC` and parsed via `big.Rat.SetString`.

## Graceful Shutdown

The server listens for `SIGINT`/`SIGTERM` via `signal.NotifyContext`. On receipt it calls
`grpc.Server.GracefulStop()` with a 15-second deadline before force-stopping.

## Structured Logging

`log/slog` is used throughout `cmd/server/main.go` with JSON output for production
compatibility with log aggregation systems.
