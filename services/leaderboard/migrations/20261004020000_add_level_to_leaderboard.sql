-- +goose Up
ALTER TABLE leaderboard_backup
    ADD COLUMN IF NOT EXISTS level INT NOT NULL DEFAULT 1;

ALTER TABLE leaderboard_backup DROP CONSTRAINT IF EXISTS leaderboard_backup_pkey;
ALTER TABLE leaderboard_backup
    ADD PRIMARY KEY (game_id, level, user_id);

DROP INDEX IF EXISTS idx_leaderboard_backup_score;
CREATE INDEX idx_leaderboard_backup_score
    ON leaderboard_backup (game_id, level, score DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_leaderboard_backup_score;
ALTER TABLE leaderboard_backup DROP CONSTRAINT IF EXISTS leaderboard_backup_pkey;
ALTER TABLE leaderboard_backup
    ADD PRIMARY KEY (game_id, user_id);
ALTER TABLE leaderboard_backup DROP COLUMN IF EXISTS level;
CREATE INDEX idx_leaderboard_backup_score ON leaderboard_backup (game_id, score DESC);
