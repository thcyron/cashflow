package api

import (
	"github.com/shopspring/decimal"

	"github.com/thcyron/cashflow/internal/cf"
)

type Transaction struct {
	Date   string `json:"date"`
	Amount string `json:"amount"`
	Shares string `json:"shares"`
	Depot  string `json:"depot"`
	Stats  Stats  `json:"stats"`
}

func encodeTransaction(transaction *cf.Transaction, stats cf.Stats) Transaction {
	return Transaction{
		Date:   transaction.Date.Format("2006-01-02"),
		Amount: transaction.Amount.String(),
		Shares: transaction.Shares.String(),
		Depot:  transaction.Depot,
		Stats:  encodeStats(stats),
	}
}

type Stats struct {
	Dividend *StatsDividend `json:"dividend"`
	Buy      *StatsBuy      `json:"buy"`
	Sell     *StatsSell     `json:"sell"`
}

type StatsDividend struct {
	Return float64 `json:"return"`
}

type StatsBuy struct {
	PricePerShare decimal.Decimal `json:"price_per_share"`
}

type StatsSell struct {
	Return        float64         `json:"return"`
	Profit        decimal.Decimal `json:"profit"`
	PricePerShare decimal.Decimal `json:"price_per_share"`
}

func encodeStats(stats cf.Stats) Stats {
	switch {
	case stats.Transaction.Shares.IsPositive():
		return Stats{
			Sell: &StatsSell{
				Return:        stats.Sell.Return,
				Profit:        stats.Sell.Profit,
				PricePerShare: stats.Sell.PricePerShare,
			},
		}
	case stats.Transaction.Shares.IsNegative():
		return Stats{
			Buy: &StatsBuy{
				PricePerShare: stats.Buy.PricePerShare,
			},
		}
	default:
		return Stats{
			Dividend: &StatsDividend{
				Return: stats.Dividend.Return,
			},
		}
	}
}
