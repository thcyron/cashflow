package cf

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

func BuildPortfolio(stocks []*Stock) Portfolio {
	p := Portfolio{}
	for _, s := range stocks {
		for _, t := range s.Transactions {
			if t.Shares.IsPositive() {
				p.RemoveShares(s, t)
			} else if t.Shares.IsNegative() {
				p.AddShares(s, t)
			} else {
				p.AddDividend(s, t)
			}
		}
	}
	return p
}

type Portfolio map[*Stock]*PortfolioStock

func (p Portfolio) Invested() decimal.Decimal {
	invested := decimal.Zero
	for _, ps := range p {
		invested = invested.Add(ps.Invested())
	}
	return invested
}

func (p Portfolio) RealizedProfit() decimal.Decimal {
	realizedProfit := decimal.Zero
	for _, ps := range p {
		realizedProfit = realizedProfit.Add(ps.RealizedProfit)
	}
	return realizedProfit
}

func (p Portfolio) Dividends() decimal.Decimal {
	dividends := decimal.Zero
	for _, ps := range p {
		dividends = dividends.Add(ps.Dividends)
	}
	return dividends
}

func (p Portfolio) AddShares(s *Stock, t *Transaction) {
	ps, ok := p[s]
	if !ok {
		ps = &PortfolioStock{}
		p[s] = ps
	}
	ps.AddShares(t)
}

func (p Portfolio) RemoveShares(s *Stock, t *Transaction) (float64, decimal.Decimal, error) {
	if p[s] == nil {
		return 0, decimal.Zero, errors.New("cf: stock not in portfolio")
	}
	return p[s].RemoveShares(t)
}

func (p Portfolio) AddDividend(s *Stock, t *Transaction) (float64, error) {
	if p[s] == nil {
		return 0, errors.New("cf: stock not in portfolio")
	}
	return p[s].AddDividend(t), nil
}

func (p Portfolio) Clone() Portfolio {
	cloned := Portfolio{}
	for s, ps := range p {
		cloned[s] = ps.Clone()
	}
	return cloned
}

type PortfolioStock struct {
	Batches        []PortfolioStockBatch
	RealizedProfit decimal.Decimal
	Dividends      decimal.Decimal
}

func (ps *PortfolioStock) Clone() *PortfolioStock {
	return &PortfolioStock{
		Batches:        append(ps.Batches[:0:0], ps.Batches...),
		RealizedProfit: ps.RealizedProfit,
		Dividends:      ps.Dividends,
	}
}

func (ps *PortfolioStock) Shares() decimal.Decimal {
	shares := decimal.Zero
	for _, b := range ps.Batches {
		shares = shares.Add(b.Shares)
	}
	return shares
}

func (ps *PortfolioStock) Invested() decimal.Decimal {
	invested := decimal.Zero
	for _, b := range ps.Batches {
		invested = invested.Add(b.Invested())
	}
	return invested
}

func (ps *PortfolioStock) PricePerShare() decimal.Decimal {
	shares := ps.Shares()
	if shares.IsZero() {
		return decimal.Zero
	}
	return ps.Invested().Div(shares)
}

func (ps *PortfolioStock) AddShares(t *Transaction) {
	ps.Batches = append(ps.Batches, PortfolioStockBatch{
		Depot:         t.Depot,
		Date:          t.Date,
		Shares:        t.Shares.Abs(),
		PricePerShare: t.Amount.Div(t.Shares),
		Transactions:  Transactions{t},
	})
}

func (ps *PortfolioStock) RemoveShares(t *Transaction) (float64, decimal.Decimal, error) {
	var (
		toRemove = t.Shares
		invested   = decimal.Zero
	)

outer:
	for toRemove.IsPositive() {
		if len(ps.Batches) == 0 {
			return 0, decimal.Zero, errors.New("cf: invalid transaction: no batches left")
		}

		for i, b := range ps.Batches {
			if b.Depot != t.Depot {
				continue
			}

			if toRemove.GreaterThanOrEqual(b.Shares) {
				toRemove = toRemove.Sub(b.Shares)
				invested = invested.Add(b.Shares.Mul(b.PricePerShare))
				ps.Batches = append(ps.Batches[0:i], ps.Batches[i+1:]...)
			} else {
				invested = invested.Add(b.PricePerShare.Mul(toRemove))
				ps.Batches[i].Shares = ps.Batches[i].Shares.Sub(toRemove)
				ps.Batches[i].Transactions = append(ps.Batches[i].Transactions, &Transaction{
					Date:   t.Date,
					Amount: t.Amount.Div(t.Shares).Mul(toRemove),
					Shares: toRemove,
					Depot:  t.Depot,
					Stock:  t.Stock,
				})
				toRemove = decimal.Zero
			}

			continue outer
		}

		return 0, decimal.Zero, errors.New("cf: invalid transaction: no batches left")
	}

	profit := t.Amount.Sub(invested)
	ps.RealizedProfit = ps.RealizedProfit.Add(profit)
	return Return(invested, t.Amount), profit, nil
}

func (ps *PortfolioStock) AddDividend(t *Transaction) float64 {
	ps.Dividends = ps.Dividends.Add(t.Amount)
	return Return(ps.Invested(), t.Amount)
}

type PortfolioStockBatch struct {
	Depot         string
	Date          time.Time
	Shares        decimal.Decimal
	PricePerShare decimal.Decimal
	Transactions  Transactions
}

func (b PortfolioStockBatch) Invested() decimal.Decimal {
	return b.Shares.Mul(b.PricePerShare)
}
