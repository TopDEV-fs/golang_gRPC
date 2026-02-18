package e2e

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	instanceadmin "cloud.google.com/go/spanner/admin/instance/apiv1"
	databaseadmin "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"

	getproduct "github.com/example/product-catalog-service/internal/app/product/queries/get_product"
	"github.com/example/product-catalog-service/internal/app/product/repo"
	activateproduct "github.com/example/product-catalog-service/internal/app/product/usecases/activate_product"
	applydiscount "github.com/example/product-catalog-service/internal/app/product/usecases/apply_discount"
	createproduct "github.com/example/product-catalog-service/internal/app/product/usecases/create_product"
	updateproduct "github.com/example/product-catalog-service/internal/app/product/usecases/update_product"
	"github.com/example/product-catalog-service/internal/pkg/clock"
	"github.com/example/product-catalog-service/internal/pkg/committer"
)

func TestProductCreationFlow(t *testing.T) {
	ctx := context.Background()
	client, dbPath := setupSpannerForTest(t, ctx)
	defer client.Close()

	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewProductReadModel(client)
	cp := committer.NewSpannerCommitter(client)
	clk := clock.NewRealClock()

	createUC := createproduct.New(productRepo, outboxRepo, cp, clk)
	getQ := getproduct.New(readModel)

	productID, err := createUC.Execute(ctx, createproduct.Request{
		Name:                 "Test Product",
		Description:          "Desc",
		Category:             "books",
		BasePriceNumerator:   1999,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	product, err := getQ.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "19.99", product.BasePrice)

	events := getOutboxEventsByAggregate(t, ctx, client, productID)
	require.NotEmpty(t, events)
	assert.Equal(t, "product.created", events[0].EventType)

	_ = dbPath
}

func TestDiscountApplicationFlow(t *testing.T) {
	ctx := context.Background()
	client, _ := setupSpannerForTest(t, ctx)
	defer client.Close()

	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewProductReadModel(client)
	cp := committer.NewSpannerCommitter(client)
	clk := clock.NewRealClock()

	createUC := createproduct.New(productRepo, outboxRepo, cp, clk)
	activateUC := activateproduct.New(productRepo, outboxRepo, cp, clk)
	discountUC := applydiscount.New(productRepo, outboxRepo, cp, clk)
	getQ := getproduct.New(readModel)

	productID, err := createUC.Execute(ctx, createproduct.Request{
		Name:                 "Discounted",
		Description:          "Desc",
		Category:             "books",
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)
	require.NoError(t, activateUC.Activate(ctx, activateproduct.ActivateRequest{ProductID: productID}))

	err = discountUC.Apply(ctx, applydiscount.ApplyRequest{
		ProductID:    productID,
		Percent:      "20",
		StartDateUTC: time.Now().Add(-time.Hour),
		EndDateUTC:   time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	product, err := getQ.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "8.00", product.EffectivePrice)
}

func TestUpdateAndStateTransitionFlow(t *testing.T) {
	ctx := context.Background()
	client, _ := setupSpannerForTest(t, ctx)
	defer client.Close()

	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo()
	cp := committer.NewSpannerCommitter(client)
	clk := clock.NewRealClock()

	createUC := createproduct.New(productRepo, outboxRepo, cp, clk)
	updateUC := updateproduct.New(productRepo, outboxRepo, cp, clk)
	activateUC := activateproduct.New(productRepo, outboxRepo, cp, clk)

	productID, err := createUC.Execute(ctx, createproduct.Request{
		Name:                 "Original",
		Description:          "Desc",
		Category:             "books",
		BasePriceNumerator:   1200,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	require.NoError(t, updateUC.Execute(ctx, updateproduct.Request{
		ProductID:   productID,
		Name:        "Updated",
		Description: "Desc 2",
		Category:    "science",
	}))
	require.NoError(t, activateUC.Activate(ctx, activateproduct.ActivateRequest{ProductID: productID}))
	require.NoError(t, activateUC.Deactivate(ctx, activateproduct.DeactivateRequest{ProductID: productID}))
}

func TestBusinessRuleValidation(t *testing.T) {
	ctx := context.Background()
	client, _ := setupSpannerForTest(t, ctx)
	defer client.Close()

	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo()
	cp := committer.NewSpannerCommitter(client)
	clk := clock.NewRealClock()

	createUC := createproduct.New(productRepo, outboxRepo, cp, clk)
	discountUC := applydiscount.New(productRepo, outboxRepo, cp, clk)

	productID, err := createUC.Execute(ctx, createproduct.Request{
		Name:                 "NoActive",
		Description:          "Desc",
		Category:             "books",
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	err = discountUC.Apply(ctx, applydiscount.ApplyRequest{
		ProductID:    productID,
		Percent:      big.NewRat(20, 1).FloatString(0),
		StartDateUTC: time.Now().Add(-time.Hour),
		EndDateUTC:   time.Now().Add(time.Hour),
	})
	assert.Error(t, err)
}

type outboxEvent struct {
	EventType string
}

func getOutboxEventsByAggregate(t *testing.T, ctx context.Context, client *spanner.Client, aggregateID string) []outboxEvent {
	t.Helper()
	stmt := spanner.Statement{SQL: `SELECT event_type FROM outbox_events WHERE aggregate_id = @id ORDER BY created_at`, Params: map[string]any{"id": aggregateID}}
	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()
	out := []outboxEvent{}
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		require.NoError(t, err)
		var eventType string
		require.NoError(t, row.Columns(&eventType))
		out = append(out, outboxEvent{EventType: eventType})
	}
	return out
}

func setupSpannerForTest(t *testing.T, ctx context.Context) (*spanner.Client, string) {
	t.Helper()
	host := os.Getenv("SPANNER_EMULATOR_HOST")
	if host == "" {
		t.Skip("SPANNER_EMULATOR_HOST is not set")
	}
	project := "local-dev"
	instanceID := "test-instance"
	dbID := "product_catalog_" + uuid.NewString()[0:8]
	instPath := "projects/" + project + "/instances/" + instanceID
	dbPath := instPath + "/databases/" + dbID

	opts := []option.ClientOption{option.WithoutAuthentication()}
	instAdmin, err := instanceadmin.NewInstanceAdminClient(ctx, opts...)
	require.NoError(t, err)
	defer instAdmin.Close()

	_, err = instAdmin.GetInstance(ctx, &instancepb.GetInstanceRequest{Name: instPath})
	if err != nil {
		op, errCreate := instAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
			Parent:     "projects/" + project,
			InstanceId: instanceID,
			Instance: &instancepb.Instance{
				Config:      "projects/" + project + "/instanceConfigs/emulator-config",
				DisplayName: "Test",
				NodeCount:   1,
			},
		})
		require.NoError(t, errCreate)
		_, err = op.Wait(ctx)
		require.NoError(t, err)
	}

	dbAdmin, err := databaseadmin.NewDatabaseAdminClient(ctx, opts...)
	require.NoError(t, err)
	defer dbAdmin.Close()

	dbOp, err := dbAdmin.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          instPath,
		CreateStatement: "CREATE DATABASE `" + dbID + "`",
		ExtraStatements: []string{readDDL(t)},
	})
	require.NoError(t, err)
	_, err = dbOp.Wait(ctx)
	require.NoError(t, err)

	client, err := spanner.NewClient(ctx, dbPath, opts...)
	require.NoError(t, err)
	return client, dbPath
}

func readDDL(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../../migrations/001_initial_schema.sql")
	require.NoError(t, err)
	return string(b)
}

