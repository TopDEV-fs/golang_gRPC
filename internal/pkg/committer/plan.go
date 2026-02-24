// Package committer provides the PlanApplier abstraction and its Spanner
// implementation. It bridges the commitplan library with the Spanner client,
// keeping use cases free from persistence-layer specifics.
package committer

import (
	"context"
	"errors"

	"cloud.google.com/go/spanner"
	"github.com/Vektor-AI/commitplan"
)

// ErrUnsupportedPlanType is returned when the commitplan.Plan implementation
// does not expose a recognised mutations accessor interface.
var ErrUnsupportedPlanType = errors.New("unsupported commitplan representation")

// PlanApplier executes a commit plan atomically against the underlying store.
type PlanApplier interface {
	// Apply commits all mutations in plan in a single read-write transaction.
	// A nil plan is treated as a no-op.
	Apply(ctx context.Context, plan *commitplan.Plan) error
}

// SpannerCommitter implements PlanApplier using a Cloud Spanner client.
type SpannerCommitter struct {
	client *spanner.Client
}

// NewSpannerCommitter returns a new SpannerCommitter backed by the given client.
func NewSpannerCommitter(client *spanner.Client) *SpannerCommitter {
	return &SpannerCommitter{client: client}
}

// Apply extracts Spanner mutations from the plan and commits them. It supports
// two mutation-access patterns exposed by different commitplan implementations.
func (c *SpannerCommitter) Apply(ctx context.Context, plan *commitplan.Plan) error {
	if plan == nil {
		return nil
	}

	if p, ok := any(plan).(interface{ Mutations() []*spanner.Mutation }); ok {
		muts := p.Mutations()
		if len(muts) == 0 {
			return nil
		}
		_, err := c.client.Apply(ctx, muts)
		return err
	}

	if p, ok := any(plan).(interface{ Items() []any }); ok {
		items := p.Items()
		muts := make([]*spanner.Mutation, 0, len(items))
		for _, it := range items {
			if m, ok := it.(*spanner.Mutation); ok && m != nil {
				muts = append(muts, m)
			}
		}
		if len(muts) == 0 {
			return nil
		}
		_, err := c.client.Apply(ctx, muts)
		return err
	}

	return ErrUnsupportedPlanType
}
