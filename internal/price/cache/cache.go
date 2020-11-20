package cache

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/thcyron/cashflow/internal/cf"
)

type Cache struct {
	provider cf.PriceProvider
	mu       sync.RWMutex
	prices   map[string][]cf.Price
}

func New(provider cf.PriceProvider) *Cache {
	return &Cache{
		provider: provider,
		prices:   map[string][]cf.Price{},
	}
}

func (c *Cache) Price(stock *cf.Stock, date time.Time) decimal.Decimal {
	c.mu.RLock()
	defer c.mu.RUnlock()

	date = cf.Date(date.Year(), int(date.Month()), date.Day())

	stockPrices := c.prices[key(stock)]
	if stockPrices == nil {
		return decimal.Zero
	}

	n := len(stockPrices)
	idx := sort.Search(n, func(i int) bool {
		return !date.Before(stockPrices[i].Date)
	})
	if idx == n {
		return decimal.Zero
	}
	return stockPrices[idx].Price
}

func (c *Cache) UpdateHistory(ctx context.Context, stocks []*cf.Stock) error {
	prices := map[string][]cf.Price{}

	for _, stock := range stocks {
		stockPrices, err := c.provider.History(ctx, stock)
		if err != nil {
			return fmt.Errorf("fetching prices for %s: %v", stock.ISIN, err)
		}
		sort.Slice(stockPrices, func(i, j int) bool {
			return stockPrices[i].Date.After(stockPrices[j].Date)
		})
		prices[key(stock)] = stockPrices
	}

	c.mu.Lock()
	c.prices = prices
	c.mu.Unlock()

	return nil
}

func key(stock *cf.Stock) string { return stock.ISIN }
