-- +goose Up
CREATE TABLE IF NOT EXISTS leaderboard_backup (
    game_id VARCHAR(50) NOT NULL,
    user_id UUID NOT NULL,
    score INT NOT NULL DEFAULT 0,
    user_email VARCHAR(255),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (game_id, user_id)
);

CREATE INDEX idx_leaderboard_backup_score ON leaderboard_backup(game_id, score DESC);

-- +goose Down
DROP TABLE IF EXISTS leaderboard_backup;