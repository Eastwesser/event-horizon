-- Module 2 / SQL: range sharding (идея для собеса, Postgres INHERITS — учебный пример).
-- На проде чаще: declarative partitioning / Citus / app-level shard key.

CREATE TABLE requests (
    id BIGSERIAL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE requests_shard_1 (
    CHECK (user_id >= 0 AND user_id < 1000000)
) INHERITS (requests);

CREATE TABLE requests_shard_2 (
    CHECK (user_id >= 1000000 AND user_id < 2000000)
) INHERITS (requests);

-- Маршрутизация insert (упрощённо через RULE; в современных версиях — PARTITION BY RANGE)
CREATE RULE requests_insert_shard_1 AS
ON INSERT TO requests
WHERE (user_id >= 0 AND user_id < 1000000)
DO INSTEAD
INSERT INTO requests_shard_1 VALUES (NEW.*);

CREATE RULE requests_insert_shard_2 AS
ON INSERT TO requests
WHERE (user_id >= 1000000 AND user_id < 2000000)
DO INSTEAD
INSERT INTO requests_shard_2 VALUES (NEW.*);

-- Что сказать: shard key = user_id; hot shard / rebalance — отдельные риски.
