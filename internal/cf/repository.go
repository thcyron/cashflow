package cf

import "context"

type Repository interface {
	Stocks(ctx context.Context) ([]*Stock, error)
}
