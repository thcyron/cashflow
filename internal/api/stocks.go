package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/thcyron/cashflow/internal/cf"
)

type Stock struct {
	Name   string  `json:"name"`
	ISIN   string  `json:"isin"`
	Symbol *string `json:"symbol"`
}

func encodeStock(stock *cf.Stock) Stock {
	encodedStock := Stock{
		Name: stock.Name,
		ISIN: stock.ISIN,
	}
	if stock.Symbol != "" {
		encodedStock.Symbol = &stock.Symbol
	}
	return encodedStock
}

type stocksResponse struct {
	Stocks []Stock `json:"stocks"`
}

func (s *Server) stocksHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	stocks, err := s.repo.Stocks(ctx)
	if err != nil {
		return fmt.Errorf("fetching stocks: %w", err)
	}

	encodedStocks := []Stock{}
	for _, stock := range stocks {
		encodedStocks = append(encodedStocks, encodeStock(stock))
	}
	return json.NewEncoder(w).Encode(stocksResponse{
		Stocks: encodedStocks,
	})
}

type stockResponse struct {
	Stock         Stock                `json:"stock"`
	Transactions  []Transaction        `json:"transactions"`
	Performances  Performances         `json:"performances"`
	Batches       []stockResponseBatch `json:"batches"`
	Invested      string               `json:"invested"`
	Value         string               `json:"value"`
	Shares        string               `json:"shares"`
	PricePerShare string               `json:"price_per_share"`
}

type stockResponseBatch struct {
	Depot         string       `json:"depot"`
	Date          string       `json:"date"`
	Shares        string       `json:"shares"`
	Invested      string       `json:"invested"`
	Value         string       `json:"value"`
	PricePerShare string       `json:"price_per_share"`
	Performances  Performances `json:"performances"`
}

func (s *Server) stockHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	stocks, err := s.repo.Stocks(ctx)
	if err != nil {
		return fmt.Errorf("fetching stocks: %w", err)
	}

	var (
		isin  = ps.ByName("isin")
		stock *cf.Stock
	)
	for _, s := range stocks {
		if s.ISIN == isin {
			stock = s
			break
		}
	}
	if stock == nil {
		http.Error(w, "stock not found", http.StatusNotFound)
		return nil
	}

	transactions, stats := cf.CalculateStats([]*cf.Stock{stock})

	encodedTransactions := []Transaction{}
	for _, transaction := range stock.Transactions {
		encodedTransactions = append(encodedTransactions, encodeTransaction(transaction, stats[transaction]))
	}

	performances := cf.CalculatePerformances(ctx, s.priceFunc, transactions, stats)

	portfolio := cf.BuildPortfolio([]*cf.Stock{stock})[stock]
	batches := []stockResponseBatch{}

	for _, batch := range portfolio.Batches {
		var (
			invested = batch.Invested()
			value    = s.priceFunc(stock, time.Now()).Mul(batch.Shares)
		)

		batchStats := batch.Transactions.Stats()
		batchPerformances := cf.CalculatePerformances(ctx, s.priceFunc, batch.Transactions, batchStats)

		batches = append(batches, stockResponseBatch{
			Depot:         batch.Depot,
			Date:          batch.Date.Format("2006-01-02"),
			Shares:        batch.Shares.String(),
			Invested:      invested.String(),
			Value:         value.String(),
			PricePerShare: batch.PricePerShare.String(),
			Performances:  EncodePerformances(batchPerformances),
		})
	}

	return json.NewEncoder(w).Encode(stockResponse{
		Stock:         encodeStock(stock),
		Transactions:  encodedTransactions,
		Performances:  EncodePerformances(performances),
		Batches:       batches,
		Invested:      portfolio.Invested().String(),
		Value:         portfolio.Invested().Add(performances.Overall.Profit).String(),
		Shares:        portfolio.Shares().String(),
		PricePerShare: portfolio.PricePerShare().String(),
	})
}
