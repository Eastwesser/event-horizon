#!/usr/bin/env python3
"""Upsert inventory card artists into authors DB (synthetic user_id UUID5).

Does not create login accounts — community/admin author counts + list only.
Never mutates Berserk card rows.

Usage (stack up):
  python3 scripts/seed_card_artists_authors.py
"""
from __future__ import annotations

import csv
import io
import subprocess
import uuid

INV_CTR = "event-horizon-postgres-inventory"
AUTH_CTR = "event-horizon-postgres-authors"


def main() -> None:
    q = """
COPY (
  SELECT DISTINCT
    COALESCE(NULLIF(attributes->>'artist_id',''), 'unknown') AS artist_id,
    COALESCE(
      NULLIF(attributes->>'artist_display',''),
      NULLIF(attributes->>'artist',''),
      attributes->>'artist_id',
      'Artist'
    ) AS display_name
  FROM items
  WHERE type = 'карточка'
    AND COALESCE(attributes->>'artist_id','') <> ''
) TO STDOUT WITH (FORMAT csv, HEADER false);
"""
    raw = subprocess.run(
        [
            "docker",
            "exec",
            "-i",
            INV_CTR,
            "psql",
            "-U",
            "eventhorizon",
            "-d",
            "eventhorizon_inventory",
            "-v",
            "ON_ERROR_STOP=1",
            "-c",
            q,
        ],
        capture_output=True,
        check=True,
    ).stdout.decode()

    rows = list(csv.reader(io.StringIO(raw)))
    if not rows:
        print("⚠️ no card artists found in inventory — seed cards first")
        return

    inserts: list[str] = []
    for artist_id, display in rows:
        artist_id = (artist_id or "").strip()
        display = (display or artist_id or "Artist").strip().replace("'", "''")
        if not artist_id or artist_id == "unknown":
            continue
        user_id = str(uuid.uuid5(uuid.NAMESPACE_URL, f"eh:artist:{artist_id}"))
        row_id = str(uuid.uuid5(uuid.NAMESPACE_URL, f"eh:author-row:{artist_id}"))
        inserts.append(
            "INSERT INTO authors (id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at) "
            f"VALUES ('{row_id}', '{user_id}', '{display}', 'Card artist (seed)', '', '', true, NOW(), NOW()) "
            "ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, updated_at = NOW();"
        )

    sql = "\n".join(inserts)
    subprocess.run(
        [
            "docker",
            "exec",
            "-i",
            AUTH_CTR,
            "psql",
            "-U",
            "eventhorizon",
            "-d",
            "eventhorizon_authors",
            "-v",
            "ON_ERROR_STOP=1",
        ],
        input=sql.encode(),
        check=True,
    )
    print(f"✅ upserted {len(inserts)} card artists into authors DB")


if __name__ == "__main__":
    main()
