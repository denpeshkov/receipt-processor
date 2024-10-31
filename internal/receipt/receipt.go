package receipt

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
)

// Item represents a single item in a [Receipt].
type Item struct {
	Description string // The product description for the item.
	Price       int64  // The total price paid for this item in cents.
}

func (i Item) String() string {
	return fmt.Sprintf("{%q $%.2f}", i.Description, float64(i.Price)/100)
}

// Receipt represents a receipt.
type Receipt struct {
	ID        string
	Retailer  string    // The name of the retailer or store the receipt is from.
	Timestamp time.Time // The timestamp of the purchase.
	Total     int64     // The total amount paid on the receipt in cents.
	Items     []Item
}

func (r Receipt) String() string {
	return fmt.Sprintf("{%q %s $%.2f %s}", r.Retailer, r.Timestamp.Format("2006-01-02 15:04"), float64(r.Total)/100, r.Items)
}

func (r *Receipt) Points() int {
	var points int

	// One point for every alphanumeric character in the retailer name.
	for _, r := range r.Retailer {
		if unicode.IsDigit(r) || unicode.IsLetter(r) {
			points++
		}
	}

	// 50 points if the total is a round dollar amount with no cents.
	if r.Total%100 == 0 {
		points += 50
	}

	// 25 points if the total is a multiple of 0.25.
	if r.Total%25 == 0 {
		points += 25
	}

	// 5 points for every two items on the receipt.
	points += 5 * (len(r.Items) / 2)

	// If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer.
	// The result is the number of points earned.
	for _, it := range r.Items {
		if len(strings.TrimSpace(it.Description))%3 == 0 {
			points += int(math.Ceil(float64(it.Price) * 0.2 / 100))
		}
	}
	// 6 points if the day in the purchase date is odd.
	if r.Timestamp.Day()%2 != 0 {
		points += 6
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	if hour := r.Timestamp.Hour(); hour >= 14 && hour < 16 {
		points += 10
	}

	return points
}
