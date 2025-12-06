CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',  
    avatar_url TEXT, 
                    -- super-admin / admin / user
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    upload_blocked BOOLEAN NOT NULL DEFAULT FALSE       -- NEW (admin control)
);

CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_banned ON users(is_banned);
CREATE INDEX idx_users_upload_blocked ON users(upload_blocked);


