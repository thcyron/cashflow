package yahoo

import (
	"context"
	"testing"
)

func TestClientDaily(t *testing.T) {
	var (
		client = NewClient()
		ctx    = context.Background()
	)

	ts, err := client.Daily(ctx, "MSFT")
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) == 0 {
		t.Fatal("no time series")
	}
	t.Logf("count: %d", len(ts))
}
