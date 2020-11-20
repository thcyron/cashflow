package cf

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type CurrentPriceFunc func(context.Context, *Stock) (decimal.Decimal, error)

type Price struct {
	Date  time.Time
	Price decimal.Decimal
}

type PriceProvider interface {
	History(ctx context.Context, stock *Stock) ([]Price, error)
	Current(ctx context.Context, stock *Stock) (decimal.Decimal, error)
}

type PriceFunc func(stock *Stock, date time.Time) decimal.Decimal

func zeroPriceFunc(stock *Stock, date time.Time) decimal.Decimal {
	return decimal.Zero
}

// func PriceFuncForStocks(ctx context.Context, stocks []*Stock, p PriceProvider) (PriceFunc, error) {
// 	g, ctx := errgroup.WithContext(ctx)
// 	results := make([][]Price, len(stocks))
// 	for i := range stocks {
// 		i := i
// 		g.Go(func() error {
// 			prices, err := p.Prices(ctx, stocks[i])
// 			if err != nil {
// 				return err
// 			}
// 			results[i] = prices
// 			return nil
// 		})
// 	}
// 	if err := g.Wait(); err != nil {
// 		return nil, err
// 	}

// 	cache := map[string][]Price{}
// 	for i := range results {
// 		stock := stocks[i]
// 		cache[stock.ISIN] = results[i]
// 	}

// 	return func(stock *Stock, date time.Time) decimal.Decimal {
// 		prices := cache[stock.ISIN]
// 		for i := len(prices) - 1; i >= 0; i-- {
// 			if p := prices[i]; !p.Date.After(date) {
// 				return p.Price
// 			}
// 		}
// 		return decimal.Zero
// 	}, nil
// }
