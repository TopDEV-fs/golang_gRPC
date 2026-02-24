package create_product_test

import (
	"context"
	"testing"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Vektor-AI/commitplan"
	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	create_product "github.com/example/product-catalog-service/internal/app/product/usecases/create_product"
	"github.com/example/product-catalog-service/internal/pkg/clock"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakeProductRepo struct {
	inserted []*domain.Product
}

func (r *fakeProductRepo) FindByID(_ context.Context, _ string) (*domain.Product, error) {
	return nil, nil
}

func (r *fakeProductRepo) InsertMut(p *domain.Product) *spanner.Mutation {
	r.inserted = append(r.inserted, p)
	return nil // nil is safe; the use case guards with != nil
}

func (r *fakeProductRepo) UpdateMut(_ *domain.Product) *spanner.Mutation { return nil }

type fakeOutboxRepo struct {
	events []contracts.OutboxEvent
}

func (o *fakeOutboxRepo) InsertMut(e contracts.OutboxEvent) *spanner.Mutation {
	o.events = append(o.events, e)
	return nil
}

type fakeCommitter struct {
	applyCalled int
	err         error
}

func (c *fakeCommitter) Apply(_ context.Context, _ *commitplan.Plan) error {
	c.applyCalled++
	return c.err
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestCreateProduct_Execute_HappyPath(t *testing.T) {
	repo := &fakeProductRepo{}
	outbox := &fakeOutboxRepo{}
	committer := &fakeCommitter{}

	uc := create_product.New(repo, outbox, committer, clock.NewRealClock())

	productID, err := uc.Execute(context.Background(), create_product.Request{
		Name:                 "Widget",
		Description:          "A fine widget",
		Category:             "hardware",
		BasePriceNumerator:   1999,
		BasePriceDenominator: 100,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, productID)
	assert.Equal(t, 1, committer.applyCalled, "plan must be applied exactly once")
	require.Len(t, repo.inserted, 1)
	assert.Equal(t, productID, repo.inserted[0].ID())
	require.Len(t, outbox.events, 1, "product.created event must be written to outbox")
	assert.Equal(t, "product.created", outbox.events[0].EventType)
}

func TestCreateProduct_Execute_InvalidName(t *testing.T) {
	uc := create_product.New(
		&fakeProductRepo{}, &fakeOutboxRepo{}, &fakeCommitter{}, clock.NewRealClock(),
	)

	_, err := uc.Execute(context.Background(), create_product.Request{
		Name:                 "   ",
		Category:             "hardware",
		BasePriceNumerator:   100,
		BasePriceDenominator: 1,
	})

	assert.ErrorIs(t, err, domain.ErrInvalidName)
}

func TestCreateProduct_Execute_InvalidPrice(t *testing.T) {
	uc := create_product.New(
		&fakeProductRepo{}, &fakeOutboxRepo{}, &fakeCommitter{}, clock.NewRealClock(),
	)

	_, err := uc.Execute(context.Background(), create_product.Request{
		Name:                 "Widget",
		Category:             "hardware",
		BasePriceNumerator:   0,
		BasePriceDenominator: 0, // denominator = 0 → ErrInvalidPrice
	})

	assert.ErrorIs(t, err, domain.ErrInvalidPrice)
}

func TestCreateProduct_Execute_PropagatesCommitterError(t *testing.T) {
	committer := &fakeCommitter{err: assert.AnError}
	uc := create_product.New(
		&fakeProductRepo{}, &fakeOutboxRepo{}, committer, clock.NewRealClock(),
	)

	_, err := uc.Execute(context.Background(), create_product.Request{
		Name:                 "Widget",
		Category:             "hardware",
		BasePriceNumerator:   100,
		BasePriceDenominator: 1,
	})

	assert.Error(t, err)
}
