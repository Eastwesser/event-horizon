#!/usr/bin/env python3
"""Seed ~10 nicknamed demo players × 8 games into leaderboard Redis (+ optional PG backup).

Usage (stack up):
  python3 scripts/seed_leaderboard_demo.py
"""
from __future__ import annotations

import subprocess
import uuid

REDIS_CTR = "event-horizon-redis-leaderboard"
PG_CTR = "event-horizon-postgres-leaderboard"
PG_DB = "eventhorizon_leaderboard"

GAMES = [
    "hexagon",
    "flappy",
    "towers",
    "hanoi",
    "memory",
    "twenty48",
    "gears",
    "companion",
]

PLAYERS = [
    ("StarPilot", 9200),
    ("NebulaKit", 8800),
    ("OrbitFox", 8400),
    ("QuarkBee", 8000),
    ("VoidMoth", 7600),
    ("IonDove", 7200),
    ("CometJay", 6800),
    ("PulsarOwl", 6400),
    ("NovaLynx", 6000),
    ("DriftWolf", 5600),
]


def redis(*args: str) -> None:
    subprocess.run(
        ["docker", "exec", REDIS_CTR, "redis-cli", *args],
        check=True,
        capture_output=True,
    )


def main() -> None:
    rows_sql: list[str] = []
    for i, (nick, base) in enumerate(PLAYERS):
        uid = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"eh.demo.lb.{nick}"))
        email = f"{nick.lower()}@demo.eventhorizon.local"
        for gi, game in enumerate(GAMES):
            score = max(10, base - gi * 120 - i * 17)
            levels = [1, 2, 3] if game == "flappy" else [1]
            for level in levels:
                key = f"leaderboard:{game}:{level}"
                ik = f"leaderboard:{game}:{level}:info"
                redis("ZADD", key, str(score), uid)
                redis("HSET", ik, f"{uid}_nickname", nick)
                redis("HSET", ik, f"{uid}_email", email)
                rows_sql.append(
                    "INSERT INTO leaderboard_backup (game_id, level, user_id, score, user_email, updated_at) "
                    f"VALUES ('{game}', {level}, '{uid}', {score}, '{email}', NOW()) "
                    "ON CONFLICT (game_id, level, user_id) DO UPDATE SET score = EXCLUDED.score, "
                    "user_email = EXCLUDED.user_email, updated_at = NOW();"
                )
        print(f"  {nick} ({uid[:8]}…)")

    sql = "\n".join(rows_sql)
    subprocess.run(
        [
            "docker",
            "exec",
            "-i",
            PG_CTR,
            "psql",
            "-U",
            "eventhorizon",
            "-d",
            PG_DB,
            "-v",
            "ON_ERROR_STOP=1",
        ],
        input=sql.encode(),
        check=True,
    )
    print(f"✅ seeded {len(PLAYERS)} players × {len(GAMES)} games into Redis + PG backup")


if __name__ == "__main__":
    main()
