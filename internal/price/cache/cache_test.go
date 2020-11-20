package cache

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/thcyron/cashflow/internal/cf"
)

type provider struct {
	prices map[string][]cf.Price
}

func (p *provider) History(ctx context.Context, stock *cf.Stock) ([]cf.Price, error) {
	return p.prices[stock.ISIN], nil
}

func (p *provider) Current(ctx context.Context, stock *cf.Stock) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func TestCache(t *testing.T) {
	provider := &provider{
		prices: map[string][]cf.Price{
			"US88160R1014": []cf.Price{
				{
					Date:  cf.Date(2020, 11, 19),
					Price: decimal.RequireFromString("499.269989"),
				},
				{
					Date:  cf.Date(2020, 11, 18),
					Price: decimal.RequireFromString("486.640015"),
				},
				{
					Date:  cf.Date(2020, 11, 20),
					Price: decimal.RequireFromString("489.609985"),
				},
				{
					Date:  cf.Date(2020, 11, 24),
					Price: decimal.RequireFromString("555.380005"),
				},
				{
					Date:  cf.Date(2020, 11, 25),
					Price: decimal.RequireFromString("574.000000"),
				},
				{
					Date:  cf.Date(2020, 11, 23),
					Price: decimal.RequireFromString("521.849976"),
				},
			},
		},
	}

	stock := &cf.Stock{ISIN: "US88160R1014"}

	cache := New(provider)
	if err := cache.UpdateHistory(context.Background(), []*cf.Stock{stock}); err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		Date  time.Time
		Price decimal.Decimal
	}{
		"too old": {
			Date:  cf.Date(2000, 1, 1),
			Price: decimal.Zero,
		},
		"exact date": {
			Date:  cf.Date(2020, 11, 23),
			Price: decimal.RequireFromString("521.849976"),
		},
		"too recent": {
			Date:  cf.Date(2020, 12, 1),
			Price: decimal.RequireFromString("574.000000"),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if price := cache.Price(stock, testCase.Date); !price.Equal(testCase.Price) {
				t.Fatalf("unexpected price: %s", price)
			}
		})
	}
}
