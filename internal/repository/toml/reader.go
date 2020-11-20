package toml

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/shopspring/decimal"

	"github.com/thcyron/cashflow/internal/cf"
)

type stockFile struct {
	Stock struct {
		Name   string
		Symbol string
		ISIN   string
	}
	Transactions []struct {
		Date   toml.LocalDate
		Amount decimal.Decimal
		Shares decimal.Decimal
		Depot  string
	} `toml:"transaction"`
}

func ReadStock(r io.Reader) (*cf.Stock, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll: %w", err)
	}

	var sf stockFile
	if err := toml.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("toml.Unmarshal: %w", err)
	}

	stock := &cf.Stock{
		Name:   sf.Stock.Name,
		Symbol: sf.Stock.Symbol,
		ISIN:   sf.Stock.ISIN,
	}
	for _, t := range sf.Transactions {
		stock.Transactions = append(stock.Transactions, &cf.Transaction{
			Date:   t.Date.In(time.UTC),
			Amount: t.Amount,
			Shares: t.Shares,
			Depot:  t.Depot,
			Stock:  stock,
		})
	}
	return stock, nil
}
