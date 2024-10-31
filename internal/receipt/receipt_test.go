package receipt

import (
	"testing"
	"time"
)

func TestPoints(t *testing.T) {
	tests := []struct {
		receipt Receipt
		points  int
	}{
		{
			receipt: Receipt{
				Retailer:  "Target",
				Timestamp: time.Date(2022, 01, 01, 13, 01, 00, 00, time.UTC),
				Items: []Item{
					{Description: "Mountain Dew 12PK", Price: 649},
					{Description: "Emils Cheese Pizza", Price: 1225},
					{Description: "Knorr Creamy Chicken", Price: 126},
					{Description: "Doritos Nacho Cheese", Price: 335},
					{Description: "Klarbrunn 12-PK 12 FL OZ  ", Price: 1200},
				},
				Total: 3535,
			},
			points: 28,
		},
		{
			receipt: Receipt{
				Retailer:  "M&M Corner Market",
				Timestamp: time.Date(2022, 03, 20, 14, 33, 00, 00, time.UTC),
				Items: []Item{
					{Description: "Gatorade", Price: 225},
					{Description: "Gatorade", Price: 225},
					{Description: "Gatorade", Price: 225},
					{Description: "Gatorade", Price: 225},
				},
				Total: 900,
			},
			points: 109,
		},
	}

	for _, tt := range tests {
		if points := tt.receipt.Points(); points != tt.points {
			t.Errorf("Points(%v) = %v, want %v", tt.receipt, points, tt.points)
		}
	}
}
