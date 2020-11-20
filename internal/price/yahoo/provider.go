package yahoo

import (
	"context"
	"errors"
	"sort"

	"github.com/shopspring/decimal"

	"github.com/thcyron/cashflow/internal/cf"
)

type Provider struct {
	client *Client
}

func NewProvider(client *Client) *Provider {
	return &Provider{
		client: client,
	}
}

func (p *Provider) History(ctx context.Context, stock *cf.Stock) ([]cf.Price, error) {
	if stock.Symbol == "" {
		return nil, errors.New("yahoo: stock is missing symbol")
	}
	ts, err := p.client.Daily(ctx, stock.Symbol)
	if err != nil {
		return nil, err
	}
	ps := make([]cf.Price, 0, len(ts))
	for d, p := range ts {
		ps = append(ps, cf.Price{
			Date:  d,
			Price: p,
		})
	}
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Date.Before(ps[j].Date)
	})
	return ps, nil
}

func (p *Provider) Current(ctx context.Context, stock *cf.Stock) (decimal.Decimal, error) {
	return p.client.Last(ctx, stock.Symbol)
}
