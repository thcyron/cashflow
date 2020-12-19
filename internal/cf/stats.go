package cf

import (
	"sort"

	"github.com/shopspring/decimal"
)

type Stats struct {
	Transaction *Transaction
	Portfolio   Portfolio
	Dividend    DividendStats
	Buy         BuyStats
	Sell        SellStats
}

type DividendStats struct {
	Return float64
}

type BuyStats struct {
	PricePerShare decimal.Decimal
}

type SellStats struct {
	Return        float64
	Profit        decimal.Decimal
	PricePerShare decimal.Decimal
}

func CalculateStats(stocks []*Stock) (Transactions, map[*Transaction]Stats, error) {
	type stockTransaction struct {
		stock *Stock
		tx    *Transaction
	}
	var sts []stockTransaction
	for _, stock := range stocks {
		for _, tx := range stock.Transactions {
			sts = append(sts, stockTransaction{
				stock: stock,
				tx:    tx,
			})
		}
	}
	sort.SliceStable(sts, func(i, j int) bool {
		return sts[i].tx.Date.Before(sts[j].tx.Date)
	})

	stats := make(map[*Transaction]Stats)
	transactions := Transactions{}

	for i, st := range sts {
		transactions = append(transactions, st.tx)

		s := Stats{
			Transaction: st.tx,
		}

		if i == 0 {
			s.Portfolio = Portfolio{}
		} else {
			s.Portfolio = stats[sts[i-1].tx].Portfolio.Clone()
		}

		if st.tx.Shares.IsPositive() {
			// Sell
			ret, profit, err := s.Portfolio.RemoveShares(st.stock, st.tx)
			if err != nil {
				return nil, nil, err
			}
			s.Sell.Return, s.Sell.Profit = ret, profit
			s.Sell.PricePerShare = st.tx.Amount.Div(st.tx.Shares)
		} else if st.tx.Shares.IsNegative() {
			// Buy
			s.Portfolio.AddShares(st.stock, st.tx)
			s.Buy.PricePerShare = st.tx.Amount.Div(st.tx.Shares)
		} else {
			// Dividend
			ret, err := s.Portfolio.AddDividend(st.stock, st.tx)
			if err != nil {
				return nil, nil, err
			}
			s.Dividend.Return = ret
		}

		stats[st.tx] = s
	}

	return transactions, stats, nil
}
