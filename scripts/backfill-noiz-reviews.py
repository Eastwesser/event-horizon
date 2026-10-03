#!/usr/bin/env python3
"""Backfill attributes.noiz_review on set-8 inventory cards from NOIZ_COMMENTS.md.

  python3 scripts/backfill-noiz-reviews.py --dry-run
  python3 scripts/backfill-noiz-reviews.py

Auth: scripts/.env.seed.admin (SEED_ADMIN_EMAIL / SEED_ADMIN_PASSWORD)
  or SEED_TOKEN / SEED_GATEWAY_URL (default http://localhost:8079)

Match: normalize names (strip paren notes, lowercase, drop punctuation)
  within set_number == 8 only. Unmatched → log + skip.
Idempotent: skip when existing noiz_review equals the parsed payload.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
import unicodedata
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Optional
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[1]
DEFAULT_COMMENTS = (
    ROOT / "confluence" / "history" / "2026-10" / "03.10.2026" / "NOIZ_COMMENTS.md"
)

SECTION_RE = re.compile(
    r"^(swamps|neutrals|mountains|woods|steppes|darkness)\b",
    re.IGNORECASE,
)
# Trailing: "9/10, имба" or "2/10 в лимитеде, 8/10 в констрактеде — скип/имба"
RATING_START_RE = re.compile(r"(\d+)\s*/\s*10\b")

@dataclass
class Review:
    source_name: str
    text: str
    rating: Optional[int]
    verdict: Optional[str]

    def payload(self) -> dict[str, Any]:
        out: dict[str, Any] = {
            "text": self.text,
            "author": "Noiz",
        }
        if self.rating is not None:
            out["rating"] = self.rating
        if self.verdict:
            out["verdict"] = self.verdict
        return out


def load_dotenv(path: Path) -> None:
    if not path.is_file():
        return
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        k, _, v = line.partition("=")
        k, v = k.strip(), v.strip().strip("'").strip('"')
        if k and k not in os.environ:
            os.environ[k] = v


def getenv(*keys: str, default: str = "") -> str:
    for k in keys:
        v = os.environ.get(k)
        if v:
            return v
    return default


def normalize_name(name: str) -> str:
    s = unicodedata.normalize("NFKC", name).strip().lower()
    s = re.sub(r"\([^)]*\)", " ", s)
    s = re.sub(r"[^\w\s\-]", " ", s, flags=re.UNICODE)
    s = re.sub(r"\s+", " ", s).strip()
    return s


def parse_tail(tail: str) -> tuple[Optional[int], Optional[str]]:
    """Parse rating/verdict from the trailing segment after the last body dash."""
    tail = tail.strip()
    if not tail:
        return None, None

    verdict: Optional[str] = None
    rating_part = tail
    if " — " in tail:
        rating_part, verdict = tail.rsplit(" — ", 1)
        verdict = verdict.strip() or None

    rm = RATING_START_RE.search(rating_part)
    rating = int(rm.group(1)) if rm else None

    if verdict is None:
        # "9/10, имба" / "6/10, норм (с потенциалом до 8)"
        sm = re.search(r"/\s*10\s*,\s*(.+)$", rating_part.strip())
        if sm:
            verdict = sm.group(1).strip() or None

    return rating, verdict


def parse_comments(path: Path) -> list[Review]:
    reviews: list[Review] = []
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line:
            continue
        if line.startswith("#") or line.startswith("карта"):
            continue
        if SECTION_RE.match(line):
            continue
        if line.startswith("Топ-") or line.startswith("Худшие"):
            continue
        if " — " not in line or "/10" not in line:
            continue

        # Body may contain em-dashes; rating segment must start with N/10.
        parts = line.split(" — ")
        if len(parts) < 3:
            continue
        name = parts[0].strip()
        # Find rightmost part that begins a rating.
        rating_idx = None
        for i in range(len(parts) - 1, 0, -1):
            if RATING_START_RE.match(parts[i].strip()):
                rating_idx = i
                break
        if rating_idx is None or rating_idx < 2:
            # rating may be "2/10 в лимитеде..." at idx, with verdict after
            for i in range(len(parts) - 1, 0, -1):
                if RATING_START_RE.search(parts[i]):
                    rating_idx = i
                    break
        if rating_idx is None or rating_idx < 2:
            print(f"SKIP parse (no rating): {line[:80]}…", file=sys.stderr)
            continue

        body = " — ".join(parts[1:rating_idx]).strip()
        tail = " — ".join(parts[rating_idx:]).strip()
        rating, verdict = parse_tail(tail)
        if not body:
            print(f"SKIP parse (empty body): {name}", file=sys.stderr)
            continue
        reviews.append(
            Review(source_name=name, text=body, rating=rating, verdict=verdict)
        )
    return reviews


def http_json(
    method: str,
    url: str,
    token: Optional[str] = None,
    body: Optional[dict] = None,
    timeout: int = 60,
) -> Any:
    data = None
    headers = {"Accept": "application/json"}
    if body is not None:
        data = json.dumps(body, ensure_ascii=False).encode("utf-8")
        headers["Content-Type"] = "application/json"
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = Request(url, data=data, headers=headers, method=method)
    try:
        with urlopen(req, timeout=timeout) as resp:
            raw = resp.read().decode("utf-8")
            return json.loads(raw) if raw else {}
    except HTTPError as e:
        err_body = e.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"HTTP {e.code} {url}: {err_body}") from e
    except URLError as e:
        raise RuntimeError(f"URL error {url}: {e}") from e


def login(base: str) -> str:
    """Prefer admin credentials.

    scripts/.env.seed.admin uses SEED_ADMIN_EMAIL/PASSWORD; author env uses
    SEED_EMAIL/PASSWORD. Preferring SEED_EMAIL first logged in as author, so
    PUT failed with 403 on admin-owned rows (e.g. Пращник) while author-owned
    cards updated fine.
    """
    token = getenv("SEED_TOKEN")
    if token:
        return token
    email = getenv(
        "SEED_ADMIN_EMAIL", "SEED_EMAIL", "SEED_AUTHOR_EMAIL",
        default="admin@eventhorizon.local",
    )
    password = getenv(
        "SEED_ADMIN_PASSWORD", "SEED_PASSWORD", "SEED_AUTHOR_PASSWORD",
        default="",
    )
    if not password:
        raise SystemExit("No SEED_TOKEN / SEED_*_PASSWORD — cannot login")
    resp = http_json(
        "POST",
        f"{base}/api/auth/login",
        body={"email": email, "password": password},
    )
    tok = resp.get("access_token") or resp.get("token")
    if not tok:
        raise SystemExit(f"login failed: {resp}")
    print(f"Logged in as {email}")
    return tok


def fetch_set8_cards(base: str, token: str) -> list[dict]:
    """All type=карточка with set_number == 8."""
    out: list[dict] = []
    offset = 0
    limit = 100
    while True:
        q = urlencode({"type": "карточка", "limit": limit, "offset": offset})
        resp = http_json("GET", f"{base}/api/inventory/items?{q}", token=token)
        batch = resp.get("items") or []
        total = int(resp.get("total") or 0)
        for it in batch:
            attrs = it.get("attributes") or {}
            sn = attrs.get("set_number")
            if sn == 8 or sn == "8":
                out.append(it)
        offset += len(batch)
        if not batch or offset >= total:
            break
    return out


def put_item(base: str, token: str, item: dict, attrs: dict) -> None:
    body = {
        "type": item.get("type") or "карточка",
        "name": item.get("name"),
        "description": item.get("description") or "",
        "price": item.get("price") or 0,
        "stock": item.get("stock") if item.get("stock") is not None else 0,
        "version": item.get("version") or 0,
        "attributes": attrs,
        "images": item.get("images") or [],
    }
    http_json(
        "PUT",
        f"{base}/api/inventory/items/{item['id']}",
        token=token,
        body=body,
    )


def main() -> int:
    load_dotenv(ROOT / "scripts" / ".env.seed.admin")
    load_dotenv(ROOT / "scripts" / ".env.seed.author")

    ap = argparse.ArgumentParser(description="Backfill Noiz reviews onto set-8 cards")
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument(
        "--comments",
        type=Path,
        default=DEFAULT_COMMENTS,
        help="Path to NOIZ_COMMENTS.md",
    )
    ap.add_argument(
        "--gateway",
        default=getenv("SEED_GATEWAY_URL", "GATEWAY_URL", default="http://localhost:8079"),
    )
    args = ap.parse_args()

    if not args.comments.is_file():
        print(f"comments file not found: {args.comments}", file=sys.stderr)
        return 1

    reviews = parse_comments(args.comments)
    by_norm: dict[str, Review] = {}
    for r in reviews:
        key = normalize_name(r.source_name)
        if not key:
            continue
        if key in by_norm:
            print(f"WARN duplicate source name key {key!r}: {r.source_name}", file=sys.stderr)
        by_norm[key] = r

    print(f"Parsed {len(reviews)} reviews ({len(by_norm)} unique keys) from {args.comments.name}")
    # Example record for audit
    sample = by_norm.get(normalize_name("Лак-нак")) or next(iter(by_norm.values()))
    print(
        "Example:",
        json.dumps(
            {
                "source_name": sample.source_name,
                "norm": normalize_name(sample.source_name),
                "noiz_review": sample.payload(),
            },
            ensure_ascii=False,
            indent=2,
        ),
    )

    if args.dry_run:
        print("--dry-run: not writing")
        return 0

    base = args.gateway.rstrip("/")
    token = login(base)
    cards = fetch_set8_cards(base, token)
    print(f"Set-8 cards in inventory: {len(cards)}")

    # Index set-8 by normalized name → list of items (foil twins etc.)
    index: dict[str, list[dict]] = {}
    for it in cards:
        key = normalize_name(it.get("name") or "")
        if key:
            index.setdefault(key, []).append(it)

    matched_keys: set[str] = set()
    updated = skipped = errors = 0
    unmatched: list[str] = []

    for key, review in sorted(by_norm.items(), key=lambda kv: kv[1].source_name):
        items = index.get(key) or []
        if not items:
            unmatched.append(review.source_name)
            continue
        matched_keys.add(key)
        payload = review.payload()
        for it in items:
            attrs = dict(it.get("attributes") or {})
            existing = attrs.get("noiz_review")
            if existing == payload:
                print(f"SKIP  {it.get('name')} [{it.get('id')}] (unchanged)")
                skipped += 1
                continue
            attrs["noiz_review"] = payload
            try:
                put_item(base, token, it, attrs)
                print(
                    f"UPDATE {it.get('name')} [{it.get('id')}] "
                    f"rating={payload.get('rating')} verdict={payload.get('verdict')!r}"
                )
                updated += 1
                it["attributes"] = attrs  # keep local cache coherent
            except Exception as e:
                print(f"ERROR {it.get('name')}: {e}", file=sys.stderr)
                errors += 1

    print("\n=== Unmatched source names (fix later) ===")
    if unmatched:
        for n in unmatched:
            print(f"  - {n}")
    else:
        print("  (none)")

    print(
        f"\nDone: reviews={len(by_norm)} matched_keys={len(matched_keys)} "
        f"updated={updated} skipped={skipped} unmatched={len(unmatched)} errors={errors}"
    )
    return 1 if errors else 0


if __name__ == "__main__":
    raise SystemExit(main())
