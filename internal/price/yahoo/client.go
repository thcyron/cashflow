package yahoo

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Daily(ctx context.Context, symbol string) (map[time.Time]decimal.Decimal, error) {
	period2 := time.Now()
	period1 := period2.AddDate(-2, 0, 0)

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v7/finance/download/%s?period1=%d&period2=%d&interval=1d&events=history",
		symbol, period1.Unix(), period2.Unix())

	prices, err := c.getPrices(ctx, url)
	if err != nil {
		return nil, err
	}

	pricesMap := map[time.Time]decimal.Decimal{}
	for _, price := range prices {
		pricesMap[price.Date] = price.Price
	}
	return pricesMap, nil
}

func (c *Client) Last(ctx context.Context, symbol string) (decimal.Decimal, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/download/%s", symbol)
	prices, err := c.getPrices(ctx, url)
	if err != nil {
		return decimal.Zero, err
	}
	if len(prices) < 2 {
		return decimal.Zero, errors.New("yahoo: no prices returned from API")
	}
	return prices[len(prices)-1].Price, nil
}

type datePrice struct {
	Date  time.Time
	Price decimal.Decimal
}

func (c *Client) getPrices(ctx context.Context, url string) ([]datePrice, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo: server responded with status code %d", resp.StatusCode)
	}

	records, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return nil, err
	}

	var prices []datePrice
	for i, record := range records {
		if i == 0 {
			continue // header
		}
		if record[4] == "null" {
			continue
		}
		date, close, err := parseRecord(record)
		if err != nil {
			return nil, err
		}
		prices = append(prices, datePrice{
			Date:  date,
			Price: close,
		})
	}
	return prices, nil
}

func parseRecord(record []string) (time.Time, decimal.Decimal, error) {
	date, err := time.Parse("2006-01-02", record[0])
	if err != nil {
		return time.Time{}, decimal.Zero, fmt.Errorf("yahoo: failed to parse date %q: %w", record[0], err)
	}
	close, err := decimal.NewFromString(record[4])
	if err != nil {
		return time.Time{}, decimal.Zero, fmt.Errorf("yahoo: failed to parse price %q: %w", record[4], err)
	}
	return date, close, nil
}
