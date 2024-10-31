package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/denpeshkov/receipt-processor/internal/receipt"
)

const (
	maxRequestBody = 1_048_576 // 1MB
)

type item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"` // USD
}

func (i *item) PriceCents() (int64, error) {
	total, err := strconv.ParseFloat(i.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("parse item price: %w", err)
	}
	return int64(total * 100), nil
}

type receiptRequest struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"` // YYYY-MM-DD
	PurchaseTime string `json:"purchaseTime"` // hh:mm
	Items        []item `json:"items"`
	Total        string `json:"total"` // USD
}

func (r *receiptRequest) PurchaseTimestamp() (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04", r.PurchaseDate+" "+r.PurchaseTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse purchase date-time: %w", err)
	}
	return t, nil
}

func (r *receiptRequest) TotalCents() (int64, error) {
	total, err := strconv.ParseFloat(r.Total, 64)
	if err != nil {
		return 0, fmt.Errorf("parse purchase total price: %w", err)
	}
	return int64(total * 100), nil
}

type receiptResponse struct {
	ID string `json:"id"`
}

func (s *Server) handleReceipt(w http.ResponseWriter, r *http.Request) error {
	var req receiptRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return &serverError{http.StatusBadRequest, fmt.Errorf("parse receipt from request: %w", err)}
	}

	pt, err := req.PurchaseTimestamp()
	if err != nil {
		return &serverError{http.StatusBadRequest, fmt.Errorf("parse receipt purchase date-time: %w", err)}
	}
	total, err := req.TotalCents()
	if err != nil {
		return &serverError{http.StatusBadRequest, fmt.Errorf("parse receipt total price: %w", err)}
	}

	items := make([]receipt.Item, len(req.Items))
	for i, it := range req.Items {
		price, err := it.PriceCents()
		if err != nil {
			return &serverError{http.StatusBadRequest, fmt.Errorf("parse item: %w", err)}
		}
		items[i] = receipt.Item{Description: it.ShortDescription, Price: price}
	}

	receipt := receipt.Receipt{
		Retailer:  req.Retailer,
		Timestamp: pt,
		Items:     items,
		Total:     total,
	}
	if err := s.storage.Store(r.Context(), &receipt); err != nil {
		return fmt.Errorf("store receipt: %w", err)
	}

	resp := receiptResponse{ID: receipt.ID}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return fmt.Errorf("encode JSON response: %w", err)
	}
	return nil
}

type pointsResponse struct {
	Points int `json:"points"`
}

func (s *Server) handlePoints(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return &serverError{http.StatusNotFound, fmt.Errorf("missing id query parameter")}
	}

	receipt, err := s.storage.Get(r.Context(), id)
	if err != nil {
		return fmt.Errorf("get receipt from storage: %w", err)
	}
	if receipt == nil {
		return &serverError{http.StatusNotFound, fmt.Errorf("receipt with id=%s not found", id)}
	}
	resp := pointsResponse{Points: receipt.Points()}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return fmt.Errorf("encode JSON response: %w", err)
	}
	return nil
}
