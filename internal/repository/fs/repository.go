package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thcyron/cashflow/internal/cf"
	"github.com/thcyron/cashflow/internal/repository/toml"
)

type Repository struct {
	dir string
}

func NewRepository(dir string) *Repository {
	return &Repository{dir: dir}
}

func (r *Repository) Stocks(ctx context.Context) ([]*cf.Stock, error) {
	return readDir(r.dir)
}

func readDir(path string) ([]*cf.Stock, error) {
	var stocks []*cf.Stock
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() && strings.HasSuffix(path, ".toml") {
			stock, err := readFile(path)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			stocks = append(stocks, stock)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stocks, nil
}

func readFile(path string) (*cf.Stock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return toml.ReadStock(f)
}
