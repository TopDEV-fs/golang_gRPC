# Architecture Notes

## Layering

- `internal/app/product/domain`: pure business rules and aggregate state transitions.
- `internal/app/product/usecases`: command orchestration using Golden Mutation Pattern.
- `internal/app/product/queries`: read-side DTO mapping and pagination.
- `internal/app/product/repo`: Spanner mapping and mutation generation.
- `internal/transport/grpc/product`: transport validation + mapping only.

## Golden Mutation Pattern

Every command use case follows:

1. Load/create aggregate.
2. Execute business method on aggregate.
3. Build `commitplan.Plan` from repository mutations.
4. Add outbox mutations for aggregate events.
5. Apply one transaction through `PlanApplier`.

## Eventing

Domain events are captured as intent structs in aggregate memory and persisted to `outbox_events` in the same transaction as aggregate writes.

## Money and Discounts

All money arithmetic uses `*big.Rat` to avoid floating-point precision errors.
