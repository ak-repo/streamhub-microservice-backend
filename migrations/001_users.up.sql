


CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    username VARCHAR(100) NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,

    role VARCHAR(20) DEFAULT 'user',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email_verified BOOLEAN DEFAULT FALSE,
    is_banned BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
