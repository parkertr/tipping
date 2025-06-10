-- Create matches read model table
CREATE TABLE IF NOT EXISTS matches_view (
    id VARCHAR(255) PRIMARY KEY,
    home_team VARCHAR(255) NOT NULL,
    away_team VARCHAR(255) NOT NULL,
    match_date TIMESTAMP NOT NULL,
    competition VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'SCHEDULED',
    home_goals INT,
    away_goals INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create predictions read model table
CREATE TABLE IF NOT EXISTS predictions_view (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    match_id VARCHAR(255) NOT NULL,
    home_goals INT NOT NULL,
    away_goals INT NOT NULL,
    points INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, match_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_matches_view_competition ON matches_view(competition);
CREATE INDEX IF NOT EXISTS idx_matches_view_match_date ON matches_view(match_date);
CREATE INDEX IF NOT EXISTS idx_matches_view_status ON matches_view(status);

CREATE INDEX IF NOT EXISTS idx_predictions_view_user_id ON predictions_view(user_id);
CREATE INDEX IF NOT EXISTS idx_predictions_view_match_id ON predictions_view(match_id);

-- Add foreign key constraints
ALTER TABLE predictions_view
ADD CONSTRAINT fk_predictions_view_match_id
FOREIGN KEY (match_id) REFERENCES matches_view(id)
ON DELETE CASCADE;
