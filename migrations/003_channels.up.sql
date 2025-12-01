-- ---------------------------------------------------------
-- Channels table (no change)
-- ---------------------------------------------------------
CREATE TABLE channels (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channels_created_by
    ON channels (created_by);

-- ---------------------------------------------------------
-- Channel members table (no change)
-- ---------------------------------------------------------
CREATE TABLE channel_members (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL,
    user_id UUID NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    UNIQUE (channel_id, user_id)
);

CREATE INDEX idx_channel_members_channel ON channel_members (channel_id);
CREATE INDEX idx_channel_members_user ON channel_members (user_id);

-- ---------------------------------------------------------
-- Messages table (add optional attachment_id)
-- ---------------------------------------------------------
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    attachment_id UUID NULL,  -- new column for optional attachment

    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY (attachment_id) REFERENCES file_attachments(id) ON DELETE SET NULL
);

CREATE INDEX idx_messages_channel_time
    ON messages (channel_id, created_at DESC);

CREATE INDEX idx_messages_sender ON messages (sender_id);

-- ---------------------------------------------------------
-- File attachments table
-- ---------------------------------------------------------
CREATE TABLE file_attachments (
    id UUID PRIMARY KEY,
    file_name TEXT NOT NULL,
    file_url TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    uploaded_by UUID NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_file_attachments_uploaded_by ON file_attachments(uploaded_by);
CREATE INDEX idx_file_attachments_file_name ON file_attachments(file_name);
