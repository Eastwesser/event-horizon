-- +goose Up
-- 3 score tiers per game (любитель / профессионал / герой) + rename Gears/Tamagotchi first_play.

UPDATE achievements
SET title = 'Первая игра в Gears',
    description = 'Сыграйте партию в Gears'
WHERE code = 'first_play_gears';

UPDATE achievements
SET title = 'Первая игра в Tamagotchi',
    description = 'Сыграйте партию в Tamagotchi'
WHERE code = 'first_play_companion';

INSERT INTO achievements (code, title, description, icon, sort_order) VALUES
    ('flappy_amateur',    'Flappy · Любитель',     'Наберите 25 очков во Flappy', 'bird', 110),
    ('flappy_pro',        'Flappy · Профессионал', 'Наберите 100 очков во Flappy', 'bird', 111),
    ('flappy_hero',       'Flappy · Герой',        'Наберите 500 очков во Flappy', 'crown', 112),
    ('hexagon_amateur',   'Pancaker · Любитель',   'Наберите 50 очков в Pancaker', 'hex', 120),
    ('hexagon_pro',       'Pancaker · Профессионал','Наберите 200 очков в Pancaker', 'hex', 121),
    ('hexagon_hero',      'Pancaker · Герой',      'Наберите 1000 очков в Pancaker', 'crown', 122),
    ('memory_amateur',    'Memonia · Любитель',    'Наберите 200 очков в Memonia', 'cards', 130),
    ('memory_pro',        'Memonia · Профессионал','Наберите 500 очков в Memonia', 'cards', 131),
    ('memory_hero',       'Memonia · Герой',       'Наберите 900 очков в Memonia', 'crown', 132),
    ('towers_amateur',    'Builder · Любитель',    'Наберите 10 очков в Builder', 'tower', 140),
    ('towers_pro',        'Builder · Профессионал','Наберите 30 очков в Builder', 'tower', 141),
    ('towers_hero',       'Builder · Герой',       'Наберите 80 очков в Builder', 'crown', 142),
    ('hanoi_amateur',     'Hanoi · Любитель',      'Наберите 100 очков в Hanoi', 'hanoi', 150),
    ('hanoi_pro',         'Hanoi · Профессионал',  'Наберите 500 очков в Hanoi', 'hanoi', 151),
    ('hanoi_hero',        'Hanoi · Герой',         'Наберите 2000 очков в Hanoi', 'crown', 152),
    ('twenty48_amateur',  '2048 · Любитель',       'Достигните 512 в 2048', 'twenty48', 160),
    ('twenty48_pro',      '2048 · Профессионал',   'Достигните 2048 в 2048', 'twenty48', 161),
    ('twenty48_hero',     '2048 · Герой',          'Достигните 8192 в 2048', 'crown', 162),
    ('gears_amateur',     'Gears · Любитель',      'Наберите 50 очков в Gears', 'gears', 170),
    ('gears_pro',         'Gears · Профессионал',  'Наберите 200 очков в Gears', 'gears', 171),
    ('gears_hero',        'Gears · Герой',         'Наберите 500 очков в Gears', 'crown', 172),
    ('companion_amateur', 'Tamagotchi · Любитель', 'Наберите 1000 очков заботы', 'star', 180),
    ('companion_pro',     'Tamagotchi · Профессионал','Наберите 5000 очков заботы', 'star', 181),
    ('companion_hero',    'Tamagotchi · Герой',    'Наберите 15000 очков заботы', 'crown', 182)
ON CONFLICT (code) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    sort_order = EXCLUDED.sort_order;

-- +goose Down
DELETE FROM user_achievements WHERE code LIKE '%_amateur' OR code LIKE '%_pro' OR code LIKE '%_hero';
DELETE FROM achievements WHERE code LIKE '%_amateur' OR code LIKE '%_pro' OR code LIKE '%_hero';
