-- Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    handle TEXT UNIQUE NOT NULL,
    title TEXT,
    description TEXT,
    avatar_url TEXT,
    seo_meta JSONB,
    theme JSONB,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for Users
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_handle ON users(handle);

-- Links Table
CREATE TABLE IF NOT EXISTS links (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    type TEXT NOT NULL,
    link_order INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB,
    click_count BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for Links
CREATE INDEX IF NOT EXISTS idx_links_user_id ON links(user_id);
-- Composite index for listing links by user and order (common query)
CREATE INDEX IF NOT EXISTS idx_links_user_order ON links(user_id, link_order);
