package xirr

import (
	"math"
	"time"
)

const (
	MaxError      = 1e-6
	MaxIterations = 100
)

type Value struct {
	Amount float64
	Date   time.Time
}

func XIRR(vs []Value, rate float64) float64 {
	if len(vs) == 0 {
		panic("xirr: no values")
	}
	for i := 0; i < MaxIterations; i++ {
		result := vs[0].Amount
		deriv := 0.0
		for j := 1; j < len(vs); j++ {
			exp := float64(vs[j].Date.Sub(vs[0].Date)/(24*time.Hour)) / 365
			result += vs[j].Amount / math.Pow(1+rate, exp)
			deriv -= exp * vs[j].Amount / math.Pow(1+rate, exp+1)
		}
		r := rate - (result / deriv)
		if math.IsNaN(r) || math.IsInf(r, 0) {
			return math.NaN()
		}
		if math.Abs(r-rate) <= MaxError || math.Abs(result) < MaxError {
			return r
		}
		rate = r
	}
	return math.NaN()
}

func Guess(vs []Value) float64 {
	if len(vs) == 0 {
		panic("xirr: no values")
	}
	var (
		debit float64
		sum   float64
	)
	for _, v := range vs {
		sum += v.Amount
		if v.Amount < 0 {
			debit += -v.Amount
		}
	}
	days := float64(vs[len(vs)-1].Date.Sub(vs[0].Date) / (time.Hour * 24))
	return math.Pow(1+(sum/debit), 365/days) - 1
}
