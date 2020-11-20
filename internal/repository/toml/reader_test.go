package toml

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"

	"github.com/thcyron/cashflow/internal/cf"
)

func TestReadStock(t *testing.T) {
	f, err := os.Open("../../../testdata/tesla.toml")
	if err != nil {
		t.Fatal(err)
	}

	stock, err := ReadStock(f)
	if err != nil {
		t.Fatal(err)
	}

	expectedStock := &cf.Stock{
		Name:   "Tesla",
		Symbol: "TSLA",
		ISIN:   "US88160R1014",
	}
	expectedStock.Transactions = []*cf.Transaction{
		{
			Date:   cf.Date(2017, 10, 6),
			Amount: decimal.RequireFromString("-3925.90"),
			Shares: decimal.RequireFromString("-25"),
			Stock:  expectedStock,
		},
		{
			Date:   cf.Date(2020, 1, 17),
			Amount: decimal.RequireFromString("1531.50"),
			Shares: decimal.RequireFromString("15"),
			Stock:  expectedStock,
		},
		{
			Date:   cf.Date(2020, 9, 23),
			Amount: decimal.RequireFromString("-3800"),
			Shares: decimal.RequireFromString("-10"),
			Stock:  expectedStock,
		},
		{
			Date:   cf.Date(2020, 10, 15),
			Amount: decimal.RequireFromString("4500"),
			Shares: decimal.RequireFromString("10"),
			Stock:  expectedStock,
		},
		{
			Date:   cf.Date(2020, 12, 9),
			Amount: decimal.RequireFromString("-7858.24"),
			Shares: decimal.RequireFromString("-13"),
			Stock:  expectedStock,
		},
		{
			Date:   cf.Date(2020, 12, 15),
			Amount: decimal.RequireFromString("-5066"),
			Shares: decimal.RequireFromString("-8"),
			Stock:  expectedStock,
		},
	}

	if !cmp.Equal(expectedStock, stock) {
		t.Fatal(cmp.Diff(expectedStock, stock))
	}
}
