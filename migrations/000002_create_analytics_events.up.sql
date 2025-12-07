-- Create analytics_events table
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL, -- 'view', 'click', 'scroll'
    link_id VARCHAR(50), -- Nullable, for link clicks
    user_id VARCHAR(50), -- Nullable, owner of the page being viewed
    visitor_id VARCHAR(100) NOT NULL, -- Fingerprint or cookie ID
    country VARCHAR(100), -- ISO country code or name
    region VARCHAR(100), -- Region/State
    meta JSONB DEFAULT '{}'::jsonb, -- Extra data: path, user_agent, referrer, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for aggregation performance
CREATE INDEX idx_analytics_event_type ON analytics_events(event_type);
CREATE INDEX idx_analytics_link_id ON analytics_events(link_id);
CREATE INDEX idx_analytics_user_id ON analytics_events(user_id);
CREATE INDEX idx_analytics_created_at ON analytics_events(created_at);
CREATE INDEX idx_analytics_visitor_id ON analytics_events(visitor_id);
