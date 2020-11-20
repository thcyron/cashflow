package cf

import (
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
	buy := decimal.Zero
	for _, pa := range p {
		buy = buy.Add(pa.Invested())
	}
	return buy
}

func (p Portfolio) RealizedProfit() decimal.Decimal {
	realizedProfit := decimal.Zero
	for _, pa := range p {
		realizedProfit = realizedProfit.Add(pa.RealizedProfit)
	}
	return realizedProfit
}

func (p Portfolio) Dividends() decimal.Decimal {
	dividends := decimal.Zero
	for _, pa := range p {
		dividends = dividends.Add(pa.Dividends)
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

func (p Portfolio) RemoveShares(s *Stock, t *Transaction) (float64, decimal.Decimal) {
	if p[s] == nil {
		panic("Portfolio.RemoveShares: not in portfolio")
	}
	return p[s].RemoveShares(t)
}

func (p Portfolio) AddDividend(s *Stock, t *Transaction) float64 {
	if p[s] == nil {
		panic("Portfolio.AddDividend: not in portfolio")
	}
	return p[s].AddDividend(t)
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
	buy := decimal.Zero
	for _, b := range ps.Batches {
		buy = buy.Add(b.Invested())
	}
	return buy
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

func (ps *PortfolioStock) RemoveShares(t *Transaction) (float64, decimal.Decimal) {
	var (
		toRemove = t.Shares
		bought   = decimal.Zero
	)

outer:
	for toRemove.IsPositive() {
		if len(ps.Batches) == 0 {
			panic("PortfolioStock.RemoveShares: no batches left")
		}

		for i, b := range ps.Batches {
			if b.Depot != t.Depot {
				continue
			}

			if toRemove.GreaterThanOrEqual(b.Shares) {
				toRemove = toRemove.Sub(b.Shares)
				bought = bought.Add(b.Shares.Mul(b.PricePerShare)) // TODO name
				ps.Batches = append(ps.Batches[0:i], ps.Batches[i+1:]...)
			} else {
				bought = bought.Add(b.PricePerShare.Mul(toRemove))
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

		panic("PortfolioStock.RemoveShares: no batches left")
	}

	ret, _ := t.Amount.Div(bought).Sub(decimal.New(1, 0)).Float64()
	profit := t.Amount.Sub(bought)
	ps.RealizedProfit = ps.RealizedProfit.Add(profit)
	return ret, profit
}

func (ps *PortfolioStock) AddDividend(t *Transaction) float64 {
	ps.Dividends = ps.Dividends.Add(t.Amount)
	buy := ps.Invested()
	if buy.IsZero() {
		return 0
	}
	ret, _ := t.Amount.Div(buy).Float64()
	return ret
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
