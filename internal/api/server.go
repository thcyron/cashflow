package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/julienschmidt/httprouter"

	"github.com/thcyron/cashflow/internal/cf"
)

type Server struct {
	logger        log.Logger
	repo          cf.Repository
	priceProvider cf.PriceProvider

	mu        sync.RWMutex
	priceFunc cf.PriceFunc
	router    *httprouter.Router
}

func New(logger log.Logger, repo cf.Repository, priceFunc cf.PriceFunc) *Server {
	s := &Server{
		logger:    logger,
		repo:      repo,
		priceFunc: priceFunc,
	}

	s.router = httprouter.New()
	s.router.GET("/stocks", s.wrap(s.stocksHandler))
	s.router.GET("/stocks/:isin", s.wrap(s.stockHandler))
	s.router.GET("/portfolio", s.wrap(s.portfolioHandler))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error

func (s *Server) wrap(handler Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		if err := handler(r.Context(), w, r, ps); err != nil {
			http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
