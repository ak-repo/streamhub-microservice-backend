-- =========================================================
-- EXTENSIONS
-- =========================================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =========================================================
-- CHANNEL STORAGE PLANS (CREATE FIRST)
-- =========================================================
CREATE TABLE channel_storage_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL UNIQUE,          -- basic / pro / enterprise
    storage_limit_mb BIGINT NOT NULL,   -- MB (precise + future-safe)
    price_inr INT NOT NULL,
    duration_days INT NOT NULL,

    CONSTRAINT chk_plan_values
        CHECK (
            storage_limit_mb > 0 AND
            price_inr >= 0 AND
            duration_days > 0
        )
);

-- =========================================================
-- CHANNELS (UPDATED)
-- =========================================================
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,

    created_by UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    description TEXT DEFAULT '',

    visibility VARCHAR(20) NOT NULL DEFAULT 'private',
    CONSTRAINT chk_visibility
        CHECK (visibility IN ('public', 'private', 'hidden')),

    is_frozen BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,

    -- ACTIVE PLAN (ID BASED)
    active_plan_id UUID NOT NULL
        REFERENCES channel_storage_plans(id),

    -- STORAGE MANAGEMENT (MB)
    storage_limit_mb BIGINT NOT NULL DEFAULT 10240,
    storage_used_mb  BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_storage_non_negative
        CHECK (
            storage_used_mb >= 0 AND
            storage_limit_mb > 0
        )
);

CREATE INDEX idx_channels_created_by   ON channels (created_by);
CREATE INDEX idx_channels_visibility   ON channels (visibility);
CREATE INDEX idx_channels_is_frozen    ON channels (is_frozen);
CREATE INDEX idx_channels_is_archived  ON channels (is_archived);
CREATE INDEX idx_channels_active_plan  ON channels (active_plan_id);

-- =========================================================
-- CHANNEL MEMBERS
-- =========================================================
CREATE TABLE channel_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    channel_id UUID NOT NULL
        REFERENCES channels(id) ON DELETE CASCADE,

    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    role TEXT NOT NULL DEFAULT 'member',

    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (channel_id, user_id)
);

CREATE INDEX idx_channel_members_channel ON channel_members (channel_id);
CREATE INDEX idx_channel_members_user    ON channel_members (user_id);

-- =========================================================
-- MESSAGES
-- =========================================================
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    channel_id UUID NOT NULL
        REFERENCES channels(id) ON DELETE CASCADE,

    sender_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_channel_time
    ON messages (channel_id, created_at DESC);

CREATE INDEX idx_messages_sender
    ON messages (sender_id);

-- =========================================================
-- CHANNEL SUBSCRIPTIONS (PAYMENT + HISTORY)
-- =========================================================
CREATE TABLE channel_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    channel_id UUID NOT NULL
        REFERENCES channels(id) ON DELETE CASCADE,

    plan_id UUID NOT NULL
        REFERENCES channel_storage_plans(id),

    razorpay_order_id TEXT NOT NULL UNIQUE,
    razorpay_payment_id TEXT UNIQUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channel_subscriptions_channel
    ON channel_subscriptions (channel_id);



CREATE INDEX idx_channel_subscriptions_order
    ON channel_subscriptions (razorpay_order_id);

