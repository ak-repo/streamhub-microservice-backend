CREATE TABLE IF NOT EXISTS channel_payment_history (
    id SERIAL PRIMARY KEY,
    channel_id UUID NOT NULL,        -- The channel receiving storage
    purchaser_user_id UUID NOT NULL, -- The admin who paid
    razorpay_session_id VARCHAR(255) UNIQUE NOT NULL,
    amount_paid_cents INT NOT NULL,
    storage_added_bytes BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);