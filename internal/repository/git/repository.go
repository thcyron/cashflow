package git

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"

	gossh "golang.org/x/crypto/ssh"

	"github.com/thcyron/cashflow/internal/cf"
	"github.com/thcyron/cashflow/internal/repository/toml"
)

const DefaultTTL = 5 * time.Minute

type Repository struct {
	TTL time.Duration

	url        string
	publicKeys *ssh.PublicKeys

	mu         sync.RWMutex
	validUntil time.Time
	stocks     []*cf.Stock
}

func NewRepository(url string) *Repository {
	return &Repository{
		TTL: DefaultTTL,
		url: url,
	}
}

func NewRepositoryWithSSH(url string, user string, pemBytes []byte) (*Repository, error) {
	publicKeys, err := ssh.NewPublicKeys(user, pemBytes, "")
	if err != nil {
		return nil, err
	}
	publicKeys.HostKeyCallback = gossh.InsecureIgnoreHostKey()
	return &Repository{
		TTL:        DefaultTTL,
		url:        url,
		publicKeys: publicKeys,
	}, nil
}

func (r *Repository) Stocks(ctx context.Context) ([]*cf.Stock, error) {
	var stocks []*cf.Stock
	r.mu.RLock()
	if r.stocks != nil && r.validUntil.After(time.Now()) {
		stocks = cloneStocks(r.stocks)
	}
	r.mu.RUnlock()
	if stocks != nil {
		return stocks, nil
	}

	stocks, err := r.getStocks(ctx)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.stocks = stocks
	r.validUntil = time.Now().Add(r.TTL)
	r.mu.Unlock()

	return cloneStocks(stocks), nil
}

func (r *Repository) getStocks(ctx context.Context) ([]*cf.Stock, error) {
	options := &git.CloneOptions{
		URL:  r.url,
		Auth: r.publicKeys,
	}
	repo, err := git.Clone(memory.NewStorage(), nil, options)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var stocks []*cf.Stock

	tree.Files().ForEach(func(f *object.File) error {
		if !strings.HasSuffix(f.Name, ".toml") {
			return nil
		}
		rc, err := f.Reader()
		if err != nil {
			return fmt.Errorf("reading %q: %w", f.Name, err)
		}
		defer rc.Close()
		stock, nil := toml.ReadStock(rc)
		if err != nil {
			return fmt.Errorf("reading %q: %w", f.Name, err)
		}
		stocks = append(stocks, stock)
		return nil
	})

	return stocks, nil
}

func cloneStocks(stocks []*cf.Stock) []*cf.Stock {
	cloned := make([]*cf.Stock, len(stocks))
	for i, stock := range stocks {
		cloned[i] = stock.Clone()
	}
	return cloned
}
