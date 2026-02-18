package repo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"github.com/example/product-catalog-service/internal/app/product/contracts"
	"github.com/example/product-catalog-service/internal/app/product/domain"
	"github.com/example/product-catalog-service/internal/models/m_outbox"
	"github.com/example/product-catalog-service/internal/models/m_product"
)

type ProductRepo struct {
	client *spanner.Client
}

func NewProductRepo(client *spanner.Client) *ProductRepo {
	return &ProductRepo{client: client}
}

func (r *ProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	stmt := spanner.Statement{
		SQL: `SELECT product_id, name, description, category, base_price_numerator, base_price_denominator,
			discount_percent, discount_start_date, discount_end_date, status, created_at, updated_at, archived_at
			FROM products WHERE product_id = @id`,
		Params: map[string]any{"id": id},
	}
	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()
	row, err := iter.Next()
	if err != nil {
		return nil, err
	}

	var (
		productID, name, description, category, status string
		baseNum, baseDen                            int64
		discountPercent                             spanner.NullNumeric
		discountStart, discountEnd                  spanner.NullTime
		createdAt, updatedAt                        time.Time
		archivedAt                                  spanner.NullTime
	)
	if err := row.Columns(&productID, &name, &description, &category, &baseNum, &baseDen, &discountPercent, &discountStart, &discountEnd, &status, &createdAt, &updatedAt, &archivedAt); err != nil {
		return nil, err
	}

	basePrice, err := domain.NewMoney(baseNum, baseDen)
	if err != nil {
		return nil, err
	}

	var discount *domain.Discount
	if discountPercent.Valid && discountStart.Valid && discountEnd.Valid {
		discount, err = domain.NewDiscount(discountPercent.Numeric.Rat, discountStart.Time, discountEnd.Time)
		if err != nil {
			return nil, err
		}
	}

	var archivedAtPtr *time.Time
	if archivedAt.Valid {
		t := archivedAt.Time
		archivedAtPtr = &t
	}

	return domain.RehydrateProduct(
		productID,
		name,
		description,
		category,
		basePrice,
		discount,
		domain.ProductStatus(status),
		createdAt,
		updatedAt,
		archivedAtPtr,
	), nil
}

func (r *ProductRepo) InsertMut(p *domain.Product) *spanner.Mutation {
	values := map[string]any{
		m_product.ProductID:            p.ID(),
		m_product.Name:                 p.Name(),
		m_product.Description:          p.Description(),
		m_product.Category:             p.Category(),
		m_product.BasePriceNumerator:   p.BasePrice().Numerator(),
		m_product.BasePriceDenominator: p.BasePrice().Denominator(),
		m_product.Status:               string(p.Status()),
		m_product.CreatedAt:            p.CreatedAt(),
		m_product.UpdatedAt:            p.UpdatedAt(),
		m_product.ArchivedAt:           nil,
	}
	if p.Discount() != nil {
		values[m_product.DiscountPercent] = spanner.Numeric{Rat: p.Discount().Percentage()}
		values[m_product.DiscountStartDate] = p.Discount().StartDate()
		values[m_product.DiscountEndDate] = p.Discount().EndDate()
	}
	return spanner.InsertMap(m_product.Table, values)
}

func (r *ProductRepo) UpdateMut(p *domain.Product) *spanner.Mutation {
	updates := map[string]any{m_product.ProductID: p.ID()}
	if p.Changes().Dirty(domain.FieldName) {
		updates[m_product.Name] = p.Name()
	}
	if p.Changes().Dirty(domain.FieldDescription) {
		updates[m_product.Description] = p.Description()
	}
	if p.Changes().Dirty(domain.FieldCategory) {
		updates[m_product.Category] = p.Category()
	}
	if p.Changes().Dirty(domain.FieldStatus) {
		updates[m_product.Status] = string(p.Status())
	}
	if p.Changes().Dirty(domain.FieldArchivedAt) {
		updates[m_product.ArchivedAt] = p.ArchivedAt()
	}
	if p.Changes().Dirty(domain.FieldDiscount) {
		if d := p.Discount(); d != nil {
			updates[m_product.DiscountPercent] = spanner.Numeric{Rat: d.Percentage()}
			updates[m_product.DiscountStartDate] = d.StartDate()
			updates[m_product.DiscountEndDate] = d.EndDate()
		} else {
			updates[m_product.DiscountPercent] = nil
			updates[m_product.DiscountStartDate] = nil
			updates[m_product.DiscountEndDate] = nil
		}
	}
	if len(updates) == 1 {
		return nil
	}
	updates[m_product.UpdatedAt] = p.UpdatedAt()
	return spanner.UpdateMap(m_product.Table, updates)
}

type OutboxRepo struct{}

func NewOutboxRepo() *OutboxRepo {
	return &OutboxRepo{}
}

func (r *OutboxRepo) InsertMut(event contracts.OutboxEvent) *spanner.Mutation {
	createdAt := time.Unix(event.CreatedAtUTC, 0).UTC()
	return spanner.InsertMap(m_outbox.Table, map[string]any{
		m_outbox.EventID:     event.EventID,
		m_outbox.EventType:   event.EventType,
		m_outbox.AggregateID: event.AggregateID,
		m_outbox.Payload:     event.Payload,
		m_outbox.Status:      event.Status,
		m_outbox.CreatedAt:   createdAt,
		m_outbox.ProcessedAt: nil,
	})
}

type ProductReadModel struct {
	client *spanner.Client
}

func NewProductReadModel(client *spanner.Client) *ProductReadModel {
	return &ProductReadModel{client: client}
}

func (rm *ProductReadModel) GetByID(ctx context.Context, id string) (*contracts.ProductDTO, error) {
	stmt := spanner.Statement{
		SQL: `SELECT product_id, name, description, category, status, base_price_numerator, base_price_denominator,
			discount_percent, discount_start_date, discount_end_date, created_at, updated_at, archived_at
			FROM products WHERE product_id = @id`,
		Params: map[string]any{"id": id},
	}
	iter := rm.client.Single().Query(ctx, stmt)
	defer iter.Stop()
	row, err := iter.Next()
	if err != nil {
		return nil, err
	}
	return scanProductDTO(row)
}

func (rm *ProductReadModel) ListActive(ctx context.Context, category string, pageSize int32, pageToken string) (*contracts.ListProductsResult, error) {
	limit := int64(pageSize)
	if limit <= 0 {
		limit = 20
	}
	offset := int64(0)
	if pageToken != "" {
		o, err := strconv.ParseInt(pageToken, 10, 64)
		if err == nil {
			offset = o
		}
	}

	query := `SELECT product_id, name, description, category, status, base_price_numerator, base_price_denominator,
		discount_percent, discount_start_date, discount_end_date, created_at, updated_at, archived_at
		FROM products WHERE status = 'ACTIVE'`
	params := map[string]any{"limit": limit, "offset": offset}
	if category != "" {
		query += ` AND category = @category`
		params["category"] = category
	}
	query += ` ORDER BY created_at DESC LIMIT @limit OFFSET @offset`

	stmt := spanner.Statement{SQL: query, Params: params}
	iter := rm.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	items := make([]contracts.ProductDTO, 0, limit)
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		dto, err := scanProductDTO(row)
		if err != nil {
			return nil, err
		}
		items = append(items, *dto)
	}

	next := ""
	if int64(len(items)) == limit {
		next = fmt.Sprintf("%d", offset+limit)
	}
	return &contracts.ListProductsResult{Items: items, NextPageToken: next}, nil
}

func scanProductDTO(row *spanner.Row) (*contracts.ProductDTO, error) {
	var (
		id, name, description, category, status string
		baseNum, baseDen                        int64
		discountPercent                         spanner.NullNumeric
		discountStart, discountEnd              spanner.NullTime
		createdAt, updatedAt                    time.Time
		archivedAt                              spanner.NullTime
	)
	if err := row.Columns(&id, &name, &description, &category, &status, &baseNum, &baseDen, &discountPercent, &discountStart, &discountEnd, &createdAt, &updatedAt, &archivedAt); err != nil {
		return nil, err
	}
	dto := &contracts.ProductDTO{
		ID:                   id,
		Name:                 name,
		Description:          description,
		Category:             category,
		Status:               status,
		BasePriceNumerator:   baseNum,
		BasePriceDenominator: baseDen,
		CreatedAtUnix:        createdAt.Unix(),
		UpdatedAtUnix:        updatedAt.Unix(),
	}
	if discountPercent.Valid {
		dto.DiscountPercent = discountPercent.Numeric.Rat.FloatString(4)
	}
	if discountStart.Valid {
		dto.DiscountStartUnix = discountStart.Time.Unix()
	}
	if discountEnd.Valid {
		dto.DiscountEndUnix = discountEnd.Time.Unix()
	}
	if archivedAt.Valid {
		dto.ArchivedAtUnix = archivedAt.Time.Unix()
	}
	return dto, nil
}
