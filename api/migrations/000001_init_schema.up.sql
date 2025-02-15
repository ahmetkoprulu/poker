-- Create users table
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    provider INTEGER NOT NULL DEFAULT 0,
    identifier VARCHAR(255) NOT NULL,
    password_hash TEXT,
    profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_identifier_not_empty CHECK (identifier != ''),
    CONSTRAINT chk_valid_provider CHECK (provider >= 0)
);

-- Create indexes for user authentication
CREATE UNIQUE INDEX idx_users_auth ON users(provider, identifier);
CREATE INDEX idx_users_provider ON users(provider);
CREATE INDEX idx_users_email_auth ON users(identifier) WHERE provider = 1;
CREATE INDEX idx_users_profile ON users USING gin (profile);

-- Add column comments
COMMENT ON COLUMN users.provider IS 'Authentication provider type: 0=None, 1=Email, 2=Google, 3=Facebook, 4=Apple';
COMMENT ON COLUMN users.identifier IS 'Provider-specific identifier: email for Email provider, provider user ID for social logins';

-- Create players table
CREATE TABLE players (
    id VARCHAR(10) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    username VARCHAR(50),
    profile_pic_url TEXT,
    chips BIGINT NOT NULL DEFAULT 1000000,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for user-player relationship
CREATE UNIQUE INDEX idx_players_user_id ON players(user_id);
