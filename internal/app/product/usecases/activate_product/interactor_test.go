package activate_product_test

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Vektor-AI/commitplan"
	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	activate_product "github.com/example/product-catalog-service/internal/app/product/usecases/activate_product"
	"github.com/example/product-catalog-service/internal/pkg/clock"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakeProductRepo struct {
	product *domain.Product
	findErr error
}

func (r *fakeProductRepo) FindByID(_ context.Context, _ string) (*domain.Product, error) {
	return r.product, r.findErr
}

func (r *fakeProductRepo) InsertMut(_ *domain.Product) *spanner.Mutation { return nil }
func (r *fakeProductRepo) UpdateMut(_ *domain.Product) *spanner.Mutation { return nil }

type fakeOutboxRepo struct{}

func (o *fakeOutboxRepo) InsertMut(_ contracts.OutboxEvent) *spanner.Mutation { return nil }

type fakeCommitter struct {
	applyCalled int
}

func (c *fakeCommitter) Apply(_ context.Context, _ *commitplan.Plan) error {
	c.applyCalled++
	return nil
}

func newInactiveProduct(t *testing.T) *domain.Product {
	t.Helper()
	price, err := domain.NewMoney(1000, 100)
	require.NoError(t, err)
	p, err := domain.NewProduct("pid-1", "Test", "Desc", "books", price, time.Now())
	require.NoError(t, err)
	p.PullDomainEvents() // discard creation event so tests start clean
	return p
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestActivate_HappyPath(t *testing.T) {
	product := newInactiveProduct(t)
	repo := &fakeProductRepo{product: product}
	committer := &fakeCommitter{}

	uc := activate_product.New(repo, &fakeOutboxRepo{}, committer, clock.NewRealClock())

	err := uc.Activate(context.Background(), activate_product.ActivateRequest{ProductID: "pid-1"})

	require.NoError(t, err)
	assert.Equal(t, domain.ProductStatusActive, product.Status())
	assert.Equal(t, 1, committer.applyCalled)
}

func TestDeactivate_HappyPath(t *testing.T) {
	product := newInactiveProduct(t)
	// Bring to ACTIVE first so we can deactivate.
	require.NoError(t, product.Activate(time.Now()))
	product.PullDomainEvents()

	repo := &fakeProductRepo{product: product}
	committer := &fakeCommitter{}

	uc := activate_product.New(repo, &fakeOutboxRepo{}, committer, clock.NewRealClock())

	err := uc.Deactivate(context.Background(), activate_product.DeactivateRequest{ProductID: "pid-1"})

	require.NoError(t, err)
	assert.Equal(t, domain.ProductStatusInactive, product.Status())
	assert.Equal(t, 1, committer.applyCalled)
}

func TestActivate_ProductNotFound(t *testing.T) {
	repo := &fakeProductRepo{findErr: assert.AnError}
	uc := activate_product.New(repo, &fakeOutboxRepo{}, &fakeCommitter{}, clock.NewRealClock())

	err := uc.Activate(context.Background(), activate_product.ActivateRequest{ProductID: "missing"})

	assert.Error(t, err)
}
