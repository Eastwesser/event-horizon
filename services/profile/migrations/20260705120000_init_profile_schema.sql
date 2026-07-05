-- +goose Up
CREATE TABLE user_profiles (
    user_id     UUID PRIMARY KEY,
    email       TEXT NOT NULL,
    nickname    TEXT,
    total_score INT DEFAULT 0,
    best_scores JSONB DEFAULT '{}',
    lamps       INT DEFAULT 0,
    tickets     INT DEFAULT 0,
    updated_at  TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS user_profiles;