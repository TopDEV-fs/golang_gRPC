.PHONY: proto run test test-unit test-e2e cover lint vet tidy migrate docker-build

SPANNER_EMULATOR_HOST ?= localhost:9010
GCP_PROJECT           ?= local-dev
SPANNER_INSTANCE      ?= test-instance
SPANNER_DATABASE      ?= product_catalog
DATABASE_PATH         := projects/$(GCP_PROJECT)/instances/$(SPANNER_INSTANCE)/databases/$(SPANNER_DATABASE)

COVER_OUT ?= coverage.out
COVER_THRESHOLD ?= 60

# ── Code generation ────────────────────────────────────────────────────────────
proto:
	@echo "Regenerate proto stubs with your preferred protoc/buf setup"

# ── Development server ─────────────────────────────────────────────────────────
run:
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) \
	SPANNER_DATABASE=$(DATABASE_PATH) \
	go run ./cmd/server

# ── Testing ────────────────────────────────────────────────────────────────────
## Run all tests (unit + e2e, requires emulator)
test:
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) \
	SPANNER_DATABASE=$(DATABASE_PATH) \
	go test -race -count=1 ./...

## Run unit tests only (no emulator required)
test-unit:
	go test -race -count=1 $(shell go list ./... | grep -v '/tests/e2e')

## Run end-to-end tests (requires emulator)
test-e2e:
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) \
	go test -v -count=1 -timeout=120s ./tests/e2e/...

## Generate coverage report and enforce a minimum threshold
cover:
	go test -race -count=1 -coverprofile=$(COVER_OUT) \
	  $(shell go list ./... | grep -v '/tests/e2e')
	go tool cover -func=$(COVER_OUT)
	@COVERAGE=$$(go tool cover -func=$(COVER_OUT) | grep total | awk '{print $$3}' | tr -d '%'); \
	  echo "Total coverage: $${COVERAGE}%"; \
	  awk -v c="$${COVERAGE}" -v t=$(COVER_THRESHOLD) \
	    'BEGIN { if (c+0 < t+0) { print "Coverage " c "% is below " t "% threshold"; exit 1 } }'

# ── Static analysis ────────────────────────────────────────────────────────────
vet:
	go vet ./...

lint:
	golangci-lint run

# ── Dependency management ──────────────────────────────────────────────────────
tidy:
	go mod tidy
	go mod verify

# ── Database ───────────────────────────────────────────────────────────────────
migrate:
	gcloud spanner databases ddl update $(SPANNER_DATABASE) \
	  --instance=$(SPANNER_INSTANCE) \
	  --project=$(GCP_PROJECT) \
	  --ddl-file=./migrations/001_initial_schema.sql

# ── Container ──────────────────────────────────────────────────────────────────
docker-build:
	docker build -t product-catalog-service:local .

