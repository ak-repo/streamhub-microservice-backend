CREATE TABLE channels (
    channel_id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for created_by
CREATE INDEX idx_channels_created_by
    ON channels (created_by);



CREATE TABLE channel_members (
    id VARCHAR(36) PRIMARY KEY,
    channel_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE,
    UNIQUE (channel_id, user_id)
);

-- Indexes
CREATE INDEX idx_channel_members_channel
    ON channel_members (channel_id);

CREATE INDEX idx_channel_members_user
    ON channel_members (user_id);




CREATE TABLE messages (
    message_id VARCHAR(36) PRIMARY KEY,
    channel_id VARCHAR(36) NOT NULL,
    sender_id VARCHAR(36) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE
);

-- Index for fast channel message history
CREATE INDEX idx_messages_channel_time
    ON messages (channel_id, created_at DESC);

-- Index for sender lookups
CREATE INDEX idx_messages_sender
    ON messages (sender_id);
