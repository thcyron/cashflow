package cf

import (
	"context"
	"math"
	"time"

	"github.com/shopspring/decimal"

	"github.com/thcyron/cashflow/internal/xirr"
)

type Performances struct {
	Overall Performance
	YTD     Performance
	Today   Performance
	IRR     float64
}

type Performance struct {
	Return float64
	Profit decimal.Decimal
}

func CalculatePerformances(ctx context.Context, price PriceFunc, transactions Transactions, stats map[*Transaction]Stats) Performances {
	if len(stats) == 0 {
		return Performances{}
	}
	var (
		now   = time.Now()
		today = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		jan1  = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	)
	return Performances{
		Overall: CalculatePerformance(ctx, price, transactions, stats, transactions[0].Date, today),
		YTD:     CalculatePerformance(ctx, price, transactions, stats, jan1, today),
		Today:   CalculatePerformance(ctx, price, transactions, stats, today, today),
		IRR:     CalculateIRR(ctx, price, transactions, stats, transactions[0].Date, today),
	}
}

func CalculatePerformance(ctx context.Context, price PriceFunc, transactions Transactions, stats map[*Transaction]Stats, begin, end time.Time) Performance {
	if len(stats) == 0 {
		return Performance{}
	}

	_, b, ok := selectTransactions(transactions, begin, end)
	if !ok {
		return Performance{}
	}

	portfolio := stats[transactions[b-1]].Portfolio
	if portfolio.Invested().IsZero() {
		return Performance{}
	}

	var (
		value     = portfolioValue(price, portfolio, end)
		invested  = decimal.Zero
		dayBefore = begin.AddDate(0, 0, -1)
	)
	for s, p := range portfolio {
		for _, b := range p.Batches {
			if b.Date.Before(begin) {
				invested = invested.Add(price(s, dayBefore).Mul(b.Shares))
			} else {
				invested = invested.Add(b.PricePerShare.Mul(b.Shares))
			}
		}
	}

	return Performance{
		Return: Return(invested, value),
		Profit: value.Sub(invested),
	}
}

func CalculateIRR(ctx context.Context, price PriceFunc, transactions Transactions, stats map[*Transaction]Stats, begin, end time.Time) float64 {
	a, b, ok := selectTransactions(transactions, begin, end)
	if !ok {
		return 0
	}

	var values []xirr.Value

	if a > 0 {
		beginningPortfolio := stats[transactions[a-1]].Portfolio
		if invested := beginningPortfolio.Invested(); invested.IsPositive() {
			values = append(values, xirr.Value{
				Date:   begin,
				Amount: -Float64(portfolioValue(price, beginningPortfolio, begin.AddDate(0, 0, -1))),
			})
		}
	}

	for _, t := range transactions[a:b] {
		values = append(values, xirr.Value{
			Date:   t.Date,
			Amount: Float64(t.Amount),
		})
	}

	endingPortfolio := stats[transactions[b-1]].Portfolio
	if invested := endingPortfolio.Invested(); invested.IsPositive() {
		values = append(values, xirr.Value{
			Date:   end,
			Amount: Float64(portfolioValue(price, endingPortfolio, end)),
		})
	}

	r := xirr.XIRR(values, xirr.Guess(values))
	if days := days(values[0].Date, values[len(values)-1].Date); days < 365 {
		r = math.Pow(1+r, float64(days)/365) - 1
	}
	return r
}

func selectTransactions(transactions Transactions, begin, end time.Time) (int, int, bool) {
	if len(transactions) == 0 {
		return 0, 0, false
	}
	if transactions[0].Date.After(end) {
		return 0, 0, false
	}
	if transactions[len(transactions)-1].Date.Before(begin) {
		return len(transactions), len(transactions), true
	}
	var (
		a = -1 // slice start of first transaction >= begin
		b = -1 // slice end of last transaction <= end
	)
	for i, t := range transactions {
		if a == -1 && (t.Date.Equal(begin) || t.Date.After(begin)) {
			a = i
		}
		if t.Date.Before(end) || t.Date.Equal(end) {
			b = i + 1
		}
	}
	if a == -1 || b == -1 {
		panic("cf.selectTransactions: invalid a or b")
	}
	return a, b, true
}

func portfolioValue(price PriceFunc, p Portfolio, date time.Time) decimal.Decimal {
	value := decimal.Zero
	for stock, pa := range p {
		if pa.Shares().IsPositive() {
			value = value.Add(price(stock, date).Mul(pa.Shares()))
		}
	}
	return value
}

func days(a, b time.Time) int {
	if a.After(b) {
		a, b = b, a
	}
	days := -a.YearDay()
	for year := a.Year(); year < b.Year(); year++ {
		days += time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC).YearDay()
	}
	return days + b.YearDay()
}
