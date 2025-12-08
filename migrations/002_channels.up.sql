-- =========================================================
-- CHANNELS
-- =========================================================
CREATE TABLE channels (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,

    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    description TEXT DEFAULT '',

    visibility VARCHAR(20) NOT NULL DEFAULT 'private',
    CONSTRAINT chk_visibility CHECK (visibility IN ('public', 'private', 'hidden')),

    is_frozen BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channels_created_by ON channels (created_by);
CREATE INDEX idx_channels_is_frozen ON channels (is_frozen);
CREATE INDEX idx_channels_is_archived ON channels (is_archived);
CREATE INDEX idx_channels_visibility ON channels (visibility);

-- =========================================================
-- CHANNEL MEMBERS
-- =========================================================
CREATE TABLE channel_members (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    role TEXT NOT NULL DEFAULT 'member',

    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (channel_id, user_id)
);

CREATE INDEX idx_channel_members_channel ON channel_members (channel_id);
CREATE INDEX idx_channel_members_user ON channel_members (user_id);

-- =========================================================
-- MESSAGES
-- =========================================================
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_channel_time ON messages (channel_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages (sender_id);
