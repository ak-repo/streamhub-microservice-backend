package domain

import "time"

type PaymentSession struct {
	RazorpayOrderID string `json:"razorpay_order_id"`

	ChannelID       string `json:"channel_id"`
	PurchaserUserID string `json:"purchaser_user_id"`
	PlanID          string `json:"plan_id"`

	AmountINR int32 `json:"amount_inr"`

	Status string `json:"status"` // created | paid | verified | failed

	CreatedAt time.Time `json:"created_at"`
}

type PaymentHistory struct {
	ID string `json:"id"` // UUID

	ChannelID       string `json:"channel_id"`
	PurchaserUserID string `json:"purchaser_user_id"`

	PlanID string `json:"plan_id"` // FK → channel_storage_plans.id

	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`

	AmountPaidINR int32 `json:"amount_paid_inr"`

	CreatedAt time.Time `json:"created_at"`
}

type SubscriptionPlan struct {
	ID string `json:"id"` // UUID

	Name string `json:"name"` // basic / pro / enterprise

	StorageLimitMB int64 `json:"storage_limit_mb"` // MB
	PriceINR       int32 `json:"price_inr"`        // INR
	DurationDays   int32 `json:"duration_days"`
}

type ChannelSubscription struct {
	ChannelID string `json:"channel_id"`

	ActivePlanID string `json:"active_plan_id"`

	StorageLimitMB int64 `json:"storage_limit_mb"`
	StorageUsedMB  int64 `json:"storage_used_mb"`

	StartedAt time.Time  `json:"started_at"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
}
