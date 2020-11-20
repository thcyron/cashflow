package cf

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	Date   time.Time
	Amount decimal.Decimal
	Shares decimal.Decimal
	Depot  string
	Stock  *Stock
}

func (t *Transaction) Clone() *Transaction {
	cloned := &Transaction{}
	*cloned = *t
	return cloned
}

type Transactions []*Transaction

func (ts Transactions) ForDepot(depot string) Transactions {
	var transactions []*Transaction
	for _, t := range ts {
		if t.Depot == depot {
			transactions = append(transactions, t)
		}
	}
	return transactions
}

func (ts Transactions) Clone() Transactions {
	cloned := make(Transactions, len(ts))
	for i, t := range ts {
		cloned[i] = t.Clone()
	}
	return cloned
}

func (ts Transactions) Stats() map[*Transaction]Stats {
	stats := make(map[*Transaction]Stats)

	for i, t := range ts {
		s := Stats{
			Transaction: t,
		}

		if i == 0 {
			s.Portfolio = Portfolio{}
		} else {
			s.Portfolio = stats[ts[i-1]].Portfolio.Clone()
		}

		if t.Shares.IsPositive() {
			// Sell
			s.Sell.Return, s.Sell.Profit = s.Portfolio.RemoveShares(t.Stock, t)
			s.Sell.PricePerShare = t.Amount.Div(t.Shares)
		} else if t.Shares.IsNegative() {
			// Buy
			s.Portfolio.AddShares(t.Stock, t)
			s.Buy.PricePerShare = t.Amount.Div(t.Shares)
		} else {
			// Dividend
			s.Dividend.Return = s.Portfolio.AddDividend(t.Stock, t)
		}

		stats[t] = s
	}

	return stats
}
