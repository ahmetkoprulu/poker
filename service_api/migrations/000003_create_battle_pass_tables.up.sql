-- Create battle_passes table
CREATE TABLE IF NOT EXISTS battle_passes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL,
    max_level INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (status IN ('active', 'expired', 'upcoming')),
    CHECK (max_level > 0),
    CHECK (start_time < end_time)
);

-- Create battle_pass_levels table
CREATE TABLE IF NOT EXISTS battle_pass_levels (
    id VARCHAR(36) PRIMARY KEY,
    battle_pass_id VARCHAR(36) NOT NULL REFERENCES battle_passes(id),
    level INTEGER NOT NULL,
    required_xp INTEGER NOT NULL,
    free_rewards JSONB NOT NULL DEFAULT '[]',
    premium_rewards JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(battle_pass_id, level),
    CHECK (level > 0),
    CHECK (required_xp > 0)
);

-- Create player_battle_passes table
CREATE TABLE IF NOT EXISTS player_battle_passes (
    id VARCHAR(36) PRIMARY KEY,
    player_id VARCHAR(36) NOT NULL,
    battle_pass_id VARCHAR(36) NOT NULL REFERENCES battle_passes(id),
    current_level INTEGER NOT NULL DEFAULT 1,
    current_xp INTEGER NOT NULL DEFAULT 0,
    is_premium BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, battle_pass_id),
    CHECK (current_level > 0),
    CHECK (current_xp >= 0)
);

-- Create player_battle_pass_rewards table
CREATE TABLE IF NOT EXISTS player_battle_pass_rewards (
    id VARCHAR(36) PRIMARY KEY,
    player_battle_pass_id VARCHAR(36) NOT NULL REFERENCES player_battle_passes(id),
    level INTEGER NOT NULL,
    is_premium BOOLEAN NOT NULL,
    claimed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_battle_pass_id, level, is_premium)
);

-- Create battle_pass_xp_transactions table
CREATE TABLE IF NOT EXISTS battle_pass_xp_transactions (
    id VARCHAR(36) PRIMARY KEY,
    player_battle_pass_id VARCHAR(36) NOT NULL REFERENCES player_battle_passes(id),
    amount INTEGER NOT NULL,
    source VARCHAR(50) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (amount > 0)
);

-- Create indexes
CREATE INDEX idx_battle_passes_status ON battle_passes(status);
CREATE INDEX idx_player_battle_passes_player_id ON player_battle_passes(player_id);
CREATE INDEX idx_battle_pass_xp_transactions_player_battle_pass_id ON battle_pass_xp_transactions(player_battle_pass_id);
CREATE INDEX idx_player_battle_pass_rewards_player_battle_pass_id ON player_battle_pass_rewards(player_battle_pass_id); 