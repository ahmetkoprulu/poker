CREATE TABLE challenges (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    category VARCHAR(50) NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    requirements JSONB NOT NULL,
    rewards JSONB NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE player_challenges (
    id UUID PRIMARY KEY,
    player_id VARCHAR(10) NOT NULL,
    challenge_id UUID NOT NULL,
    progress INT NOT NULL DEFAULT 0,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    reward_claimed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (player_id) REFERENCES players(id),
    FOREIGN KEY (challenge_id) REFERENCES challenges(id)
);

CREATE INDEX idx_player_challenges_player ON player_challenges(player_id);
CREATE INDEX idx_active_challenges ON player_challenges(player_id, completed) WHERE NOT completed;