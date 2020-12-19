package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/thcyron/cashflow/internal/cf"
)

type Portfolio struct {
	Stocks       []PortfolioStock `json:"stocks"`
	Invested     string           `json:"invested"`
	Value        string           `json:"value"`
	Performances Performances     `json:"performances"`
}

type PortfolioStock struct {
	Stock         Stock                 `json:"stock"`
	Batches       []PortfolioStockBatch `json:"batches"`
	Invested      string                `json:"invested"`
	Value         string                `json:"value"`
	Shares        string                `json:"shares"`
	PricePerShare string                `json:"price_per_share"`
	Performances  Performances          `json:"performances"`
}

type PortfolioStockBatch struct {
	Depot         string       `json:"depot"`
	Date          string       `json:"date"`
	Shares        string       `json:"shares"`
	PricePerShare string       `json:"price_per_share"`
	Invested      string       `json:"invested"`
	Value         string       `json:"value"`
	Performances  Performances `json:"performances"`
}

type Performances struct {
	Overall Performance `json:"overall"`
	YTD     Performance `json:"ytd"`
	Today   Performance `json:"today"`
	IRR     *float64    `json:"irr"`
}

func EncodePerformances(performances cf.Performances) Performances {
	return Performances{
		Overall: EncodePerformance(performances.Overall),
		YTD:     EncodePerformance(performances.YTD),
		Today:   EncodePerformance(performances.Today),
		IRR:     encodeReturn(performances.IRR),
	}
}

type Performance struct {
	Return *float64 `json:"return"`
	Profit string   `json:"profit"`
}

func EncodePerformance(performance cf.Performance) Performance {
	return Performance{
		Return: encodeReturn(performance.Return),
		Profit: performance.Profit.String(),
	}
}

func encodeReturn(ret float64) *float64 {
	if !math.IsNaN(ret) && !math.IsInf(ret, 0) {
		return &ret
	}
	return nil
}

func EncodePortfolioStock(stock *cf.Stock, portfolioStock *cf.PortfolioStock) PortfolioStock {
	encodedStock := PortfolioStock{
		Stock:         encodeStock(stock),
		Batches:       []PortfolioStockBatch{},
		Invested:      portfolioStock.Invested().String(),
		Shares:        portfolioStock.Shares().String(),
		PricePerShare: portfolioStock.PricePerShare().String(),
	}
	for _, batch := range portfolioStock.Batches {
		encodedStock.Batches = append(encodedStock.Batches, EncodePortfolioStockBatch(batch))
	}
	return encodedStock
}

func EncodePortfolioStockBatch(batch cf.PortfolioStockBatch) PortfolioStockBatch {
	return PortfolioStockBatch{
		Depot:         batch.Depot,
		Date:          batch.Date.Format("2006-01-02"),
		Shares:        batch.Shares.String(),
		PricePerShare: batch.PricePerShare.String(),
		Invested:      batch.Invested().String(),
	}
}

type portfolioResponse struct {
	Portfolio Portfolio `json:"portfolio"`
}

func (s *Server) portfolioHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	stocks, err := s.repo.Stocks(ctx)
	if err != nil {
		return fmt.Errorf("fetching stocks: %w", err)
	}

	if symbol := r.URL.Query().Get("stock"); symbol != "" {
		var found *cf.Stock
		for _, stock := range stocks {
			if stock.Symbol == symbol {
				found = stock
				break
			}
		}
		if found != nil {
			stocks = []*cf.Stock{found}
		} else {
			stocks = []*cf.Stock{}
		}
	}

	transactions, stats, err := cf.CalculateStats(stocks)
	if err != nil {
		return err
	}

	portfolio := cf.BuildPortfolio(stocks)
	performances := cf.CalculatePerformances(ctx, s.priceFunc, transactions, stats)

	encodedPortfolio := Portfolio{
		Stocks:       []PortfolioStock{},
		Invested:     portfolio.Invested().String(),
		Value:        portfolio.Invested().Add(performances.Overall.Profit).String(),
		Performances: EncodePerformances(performances),
	}
	for stock, portfolioStock := range portfolio {
		if portfolioStock.Shares().IsPositive() {
			encodedPortfolioStock := EncodePortfolioStock(stock, portfolioStock)

			stockTransactions, stockStats, err := cf.CalculateStats([]*cf.Stock{stock})
			if err != nil {
				return err
			}

			performances := cf.CalculatePerformances(ctx, s.priceFunc, stockTransactions, stockStats)
			encodedPortfolioStock.Performances = EncodePerformances(performances)

			encodedPortfolioStock.Value = portfolioStock.Invested().Add(performances.Overall.Profit).String()
			encodedPortfolio.Stocks = append(encodedPortfolio.Stocks, encodedPortfolioStock)
		}
	}

	return json.NewEncoder(w).Encode(portfolioResponse{
		Portfolio: encodedPortfolio,
	})
}
