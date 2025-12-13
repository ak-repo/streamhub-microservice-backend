package domain

import (
	"time"
)

// PaymentHistory represents a record in the channel_payment_history table.
type PaymentHistory struct {
	ID                string // UUID
	ChannelID         string
	PurchaserUserID   string
	RazorpayOrderID   string
	RazorpayPaymentID string
	AmountPaidCents   int32
	StorageAddedBytes int64
	CreatedAt         time.Time
}

type PaymentSession struct {
	RazorpayOrderID string    `json:"razorpay_order_id"`
	ChannelID       string    `json:"channel_id"`
	PurchaserUserID string    `json:"user_id"`
	AmountCents     int32     `json:"amount_cents"`
	StorageBytes    int64     `json:"storage_bytes"`
	Status          string    `json:"status"` // created | paid | verified | failed
	CreatedAt       time.Time `json:"created_at"`
}
