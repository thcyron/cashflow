package xirr

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestXIRR(t *testing.T) {
	testCases := []struct {
		Values []Value
		Rate   float64
	}{
		{
			Values: []Value{
				{-1000, time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
				{-123, time.Date(2018, 1, 10, 0, 0, 0, 0, time.UTC)},
				{1130, time.Date(2018, 2, 3, 0, 0, 0, 0, time.UTC)},
			},
			Rate: 0.07341451967,
		},
		{
			Values: []Value{
				{-1000, time.Date(2015, 6, 11, 0, 0, 0, 0, time.UTC)},
				{-9000, time.Date(2015, 7, 21, 0, 0, 0, 0, time.UTC)},
				{-3000, time.Date(2015, 10, 17, 0, 0, 0, 0, time.UTC)},
				{20000, time.Date(2018, 6, 10, 0, 0, 0, 0, time.UTC)},
			},
			Rate: 0.1635371584432641,
		},
		{
			Values: []Value{
				{-9997, time.Date(2019, 10, 23, 0, 0, 0, 0, time.UTC)},
				{9662, time.Date(2019, 10, 25, 0, 0, 0, 0, time.UTC)},
			},
			Rate: -0.998012,
		},
		{
			Values: []Value{
				{-1809.45, time.Date(2016, 10, 10, 0, 0, 0, 0, time.UTC)},
				{-1878.96, time.Date(2016, 11, 4, 0, 0, 0, 0, time.UTC)},
				{3587.577, time.Date(2016, 11, 4, 0, 0, 0, 0, time.UTC)},
			},
			Rate: -0.567055,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", i+1), func(t *testing.T) {
			r := XIRR(testCase.Values, Guess(testCase.Values))
			if math.IsNaN(r) {
				t.Fatal("NaN")
			}
			if e := math.Abs(r - testCase.Rate); e > MaxError {
				t.Fatalf("unexpected rate: %f", r)
			}
		})
	}
}
