package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/denpeshkov/receipt-processor/internal/receipt"
)

func TestReceiptRequest_PurchaseTimestamp(t *testing.T) {
	tests := []struct {
		date, time string
		ts         time.Time
	}{
		{
			date: "2022-01-01", time: "13:01",
			ts: time.Date(2022, 01, 01, 13, 01, 00, 00, time.UTC),
		},
		{
			date: "2022-03-20", time: "14:33",
			ts: time.Date(2022, 03, 20, 14, 33, 00, 00, time.UTC),
		},
	}

	for _, tt := range tests {
		req := receiptRequest{PurchaseDate: tt.date, PurchaseTime: tt.time}
		ts, err := req.PurchaseTimestamp()
		if err != nil {
			t.Fatalf("PurchaseTimestamp(%s %s) failed: %v", tt.date, tt.time, err)
		}
		if ts != tt.ts {
			t.Errorf("PurchaseTimestamp(%s %s) = %s, want %s", tt.date, tt.time, ts.String(), tt.ts.String())
		}
	}
}

func TestReceiptRequest_TotalCents(t *testing.T) {
	tests := []struct {
		usd   string
		cents int64
	}{
		{usd: "6.49", cents: 649},
		{usd: "12.25", cents: 1225},
		{usd: "1.26", cents: 126},
		{usd: "3.35", cents: 335},
		{usd: "12.00", cents: 1200},
		{usd: "2.25", cents: 225},
		{usd: "12345.67", cents: 1234567},
		{usd: "0.01", cents: 1},
		{usd: "0.99", cents: 99},
		{usd: "0", cents: 0},
		{usd: "7", cents: 700},
	}

	for _, tt := range tests {
		req := receiptRequest{Total: tt.usd}
		cents, err := req.TotalCents()
		if err != nil {
			t.Fatalf("TotalCents(%s) failed: %v", tt.usd, err)
		}
		if cents != tt.cents {
			t.Errorf("TotalCents(%s) = %d, want %d", tt.usd, cents, tt.cents)
		}
	}
}

type mockStorage struct {
	store func(context.Context, *receipt.Receipt) error
	get   func(context.Context, string) (*receipt.Receipt, error)
}

func (m mockStorage) Store(ctx context.Context, r *receipt.Receipt) error {
	return m.store(ctx, r)
}
func (m mockStorage) Get(ctx context.Context, id string) (*receipt.Receipt, error) {
	return m.get(ctx, id)
}

func TestHandleReceipt(t *testing.T) {
	storage := mockStorage{
		store: func(context.Context, *receipt.Receipt) error { return nil },
		get:   func(context.Context, string) (*receipt.Receipt, error) { return nil, nil },
	}
	server := httptest.NewServer(New(storage, Options{Logger: slog.Default()}))
	t.Cleanup(server.Close)

	client := server.Client()

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{
			name: "200 OK",
			body: `
			{
				"retailer": "test retailer",
				"purchaseDate": "2022-01-01",
				"purchaseTime": "13:01",
				"items": [{"shortDescription": "test item","price": "6.49"}],
				"total": "35.35"
			}`,
			status: http.StatusOK,
		},
		{
			name: "400 Bad Request",
			body: `
			{
				"retailer": 42,
				"purchaseDate": "2022-01-01",
				"purchaseTime": "13:01",
				"items": [{"shortDescription": "test item","price": "6.49"}],
				"total": "35.35"
			}`,
			status: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.Post(server.URL+"/receipts/process", "application/json", strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Post() failed: %v", err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })

			if resp.StatusCode != tt.status {
				t.Errorf("Got status %d, want %d", resp.StatusCode, tt.status)
			}
		})
	}
}

func TestHandlePoints(t *testing.T) {
	storage := mockStorage{
		store: func(context.Context, *receipt.Receipt) error { return nil },
		get: func(ctx context.Context, id string) (*receipt.Receipt, error) {
			if id == "not found" {
				return nil, nil
			}
			return &receipt.Receipt{}, nil
		},
	}
	server := httptest.NewServer(New(storage, Options{Logger: slog.Default()}))
	t.Cleanup(server.Close)

	client := server.Client()

	tests := []struct {
		name   string
		id     string
		status int
	}{
		{name: "200 OK", id: "found", status: http.StatusOK},
		{name: "404 Not Found", id: "not found", status: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.Get(fmt.Sprintf("%s/receipts/%s/points", server.URL, tt.id))
			if err != nil {
				t.Fatalf("Get() failed: %v", err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })

			if resp.StatusCode != tt.status {
				t.Errorf("Got status %d, want %d", resp.StatusCode, tt.status)
			}
		})
	}
}
