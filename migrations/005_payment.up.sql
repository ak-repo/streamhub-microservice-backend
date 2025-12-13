CREATE TABLE payment_history (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL,
    purchaser_user_id UUID NOT NULL,
    razorpay_order_id TEXT NOT NULL,
    razorpay_payment_id TEXT NOT NULL UNIQUE,
    amount_paid_cents INTEGER NOT NULL,
    storage_added_bytes BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
