-- +goose Up
ALTER TABLE highscores
    ADD COLUMN IF NOT EXISTS level INT NOT NULL DEFAULT 1;

ALTER TABLE highscores DROP CONSTRAINT IF EXISTS highscores_pkey;
ALTER TABLE highscores
    ADD PRIMARY KEY (user_id, game_id, level);

-- +goose Down
ALTER TABLE highscores DROP CONSTRAINT IF EXISTS highscores_pkey;
ALTER TABLE highscores
    ADD PRIMARY KEY (user_id, game_id);
ALTER TABLE highscores DROP COLUMN IF EXISTS level;
