.PHONY: proto run test migrate

SPANNER_EMULATOR_HOST ?= localhost:9010
GCP_PROJECT ?= local-dev
SPANNER_INSTANCE ?= test-instance
SPANNER_DATABASE ?= product_catalog
DATABASE_PATH := projects/$(GCP_PROJECT)/instances/$(SPANNER_INSTANCE)/databases/$(SPANNER_DATABASE)

proto:
	@echo "Generate proto files with your preferred protoc setup"

run:
	set SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) && set SPANNER_DATABASE=$(DATABASE_PATH) && go run ./cmd/server

test:
	set SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) && set SPANNER_DATABASE=$(DATABASE_PATH) && go test ./...

migrate:
	gcloud spanner databases ddl update $(SPANNER_DATABASE) --instance=$(SPANNER_INSTANCE) --project=$(GCP_PROJECT) --ddl-file=./migrations/001_initial_schema.sql
