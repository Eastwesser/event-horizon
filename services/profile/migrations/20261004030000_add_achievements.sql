-- +goose Up
CREATE TABLE IF NOT EXISTS achievements (
    code        TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon        TEXT NOT NULL DEFAULT 'trophy',
    sort_order  INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_achievements (
    user_id     UUID NOT NULL REFERENCES user_profiles(user_id) ON DELETE CASCADE,
    code        TEXT NOT NULL REFERENCES achievements(code) ON DELETE CASCADE,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, code)
);

CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements (user_id);

INSERT INTO achievements (code, title, description, icon, sort_order) VALUES
    ('first_play_flappy',    'Первая игра во Flappy Bird', 'Сыграйте партию во Flappy Bird', 'bird', 10),
    ('first_play_hexagon',   'Первая игра в Pancaker',     'Сыграйте партию в Pancaker', 'hex', 11),
    ('first_play_memory',    'Первая игра в Memonia',      'Сыграйте партию в Memonia', 'cards', 12),
    ('first_play_towers',    'Первая игра в Builder',      'Сыграйте партию в Builder', 'tower', 13),
    ('first_play_hanoi',     'Первая игра в Hanoi',        'Сыграйте партию в Hanoi', 'hanoi', 14),
    ('first_play_twenty48',  'Первая игра в 2048',         'Сыграйте партию в 2048', 'twenty48', 15),
    ('first_play_gears',     'Первая игра в Орбиты',       'Сыграйте партию в Орбиты', 'gears', 16),
    ('first_play_companion', 'Первая игра в Компаньон',    'Сыграйте партию в Компаньон', 'star', 17),
    ('flappy_score_100',     '100 очков во Flappy',        'Наберите 100 очков во Flappy Bird', 'bird', 20),
    ('flappy_score_500',     '500 очков во Flappy',        'Наберите 500 очков во Flappy Bird', 'medal', 21),
    ('flappy_level_5',       'Flappy: уровень 5',          'Достигните 5 уровня во Flappy Bird', 'star', 22),
    ('flappy_level_10',      'Flappy: уровень 10',         'Достигните 10 уровня во Flappy Bird', 'crown', 23),
    ('total_score_1k',       '1000 очков всего',           'Сумма лучших результатов ≥ 1000', 'trophy', 30),
    ('total_score_5k',       '5000 очков всего',           'Сумма лучших результатов ≥ 5000', 'sparkle', 31),
    ('total_score_10k',      '10000 очков всего',          'Сумма лучших результатов ≥ 10000', 'crown', 32)
ON CONFLICT (code) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    sort_order = EXCLUDED.sort_order;

-- +goose Down
DROP TABLE IF EXISTS user_achievements;
DROP TABLE IF EXISTS achievements;
