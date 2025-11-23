CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    original_name TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    content_type TEXT NOT NULL,

    size BIGINT NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_files_owner_id ON files(owner_id);
CREATE INDEX IF NOT EXISTS idx_files_public ON files(is_public);
