package storage

import (
	"context"
	"testing"
	"time"

	"github.com/denpeshkov/receipt-processor/internal/receipt"
)

func Test(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	r := &receipt.Receipt{
		Retailer:  "retailer name",
		Timestamp: now,
		Items: []receipt.Item{
			{Description: "item 1 name ", Price: 4200},
			{Description: "item 2 name", Price: 5137},
		},
		Total: 9337,
	}
	s := New()
	if err := s.Store(ctx, r); err != nil {
		t.Fatalf("Store(%+v) failed: %v", r, err)
	}
	r2, err := s.Get(ctx, r.ID)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", r.ID, err)
	}
	if r2 != r {
		t.Errorf("Got: %+v, want: %+v", r2, r)
	}
}
