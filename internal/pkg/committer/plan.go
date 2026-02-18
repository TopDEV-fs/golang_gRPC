package committer

import (
	"context"

	"cloud.google.com/go/spanner"
	"github.com/Vektor-AI/commitplan"
)

type PlanApplier interface {
	Apply(ctx context.Context, plan *commitplan.Plan) error
}

type SpannerCommitter struct {
	client *spanner.Client
}

func NewSpannerCommitter(client *spanner.Client) *SpannerCommitter {
	return &SpannerCommitter{client: client}
}

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

	return nil
}
