-- +goose Up
CREATE TABLE IF NOT EXISTS highscores (
    user_id UUID NOT NULL,
    game_id VARCHAR(50) NOT NULL,
    score INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, game_id)
);

CREATE INDEX idx_highscores_game_id ON highscores(game_id);
CREATE INDEX idx_highscores_score ON highscores(score);

-- +goose Down
DROP TABLE IF EXISTS highscores;