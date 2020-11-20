package cf

import (
	"time"

	"github.com/shopspring/decimal"
)

func Date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func Return(invested, value decimal.Decimal) float64 {
	if invested.IsZero() || value.IsZero() {
		return 0
	}
	return Float64(value.Div(invested)) - 1
}

func Float64(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}
