#!/usr/bin/env python3
"""Seed Berserk CCG cards into inventory from *_cards_info.md + seed/README.md stock.

  python3 scripts/seed-berserk-cards.py --dry-run --smoke
  python3 scripts/seed-berserk-cards.py --smoke
  python3 scripts/seed-berserk-cards.py --limit 5
  python3 scripts/seed-berserk-cards.py

Auth (author/admin JWT):
  SEED_GATEWAY_URL (default http://localhost:8079)
  scripts/.env.seed.admin or .env.seed.author — SEED_*_EMAIL / PASSWORD
  or SEED_EMAIL / SEED_PASSWORD / SEED_TOKEN

Pricing (CARDS_PRICING.md): market_rub × 1000 tickets; foil ×2.
  common 10→10k, uncommon 25→25k, rare 50→50k, ultra 100→100k
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
import unicodedata
from dataclasses import dataclass, field, asdict
from pathlib import Path
from typing import Any, Optional
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[1]
SEED_ROOT = ROOT / "services" / "inventory" / "internal" / "seed"
CARDS_ROOT = SEED_ROOT / "berserk_cards" / "year_set"
STOCK_README = SEED_ROOT / "README.md"

RARITY_RUB = {
    "common": 10,
    "uncommon": 25,
    "rare": 50,
    "ultra": 100,
}
TICKETS_PER_RUB = 1000

RARITY_RU = {
    "обычная": "common",
    "необычная": "uncommon",
    "редкая": "rare",
    "ультраредкая": "ultra",
    "ультра": "ultra",
}

ELEMENT_RU = {
    "горы": "mountains",
    "гора": "mountains",
    "лес": "woods",
    "леса": "woods",  # plural form used in cards_info («Стихия Леса»)
    "степи": "steppes",
    "степь": "steppes",
    "болота": "swamps",
    "болото": "swamps",
    "тьма": "darkness",
    "нейтральная": "neutral",
    "нейтралы": "neutral",
    "нейтрал": "neutral",
    # Already-canonical English codes (pass-through if present in source)
    "mountains": "mountains",
    "woods": "woods",
    "steppes": "steppes",
    "swamps": "swamps",
    "darkness": "darkness",
    "neutral": "neutral",
}

TYPE_MAIN_RU = {
    "существо": "creature",
    "артефакт": "artifact",
    "местность": "land",
}

COST_TIER_RU = {
    "рядовая": "rank_and_file",
    "элитная": "elite",
}

# Smoke cards: (normalized display name without foil suffix, set_number)
SMOKE_ORDER = [
    ("дирр", 4),
    ("тхакрай", 7),
    ("мормолика", 8),  # expands to 2 items
    ("склеп керсам", 8),
    ("беллигемин", 7),
]

# Cyrillic → latin for artist_id (simple phonetic)
_CYR = {
    "а": "a", "б": "b", "в": "v", "г": "g", "д": "d", "е": "e", "ё": "e",
    "ж": "zh", "з": "z", "и": "i", "й": "y", "к": "k", "л": "l", "м": "m",
    "н": "n", "о": "o", "п": "p", "р": "r", "с": "s", "т": "t", "у": "u",
    "ф": "f", "х": "kh", "ц": "ts", "ч": "ch", "ш": "sh", "щ": "shch",
    "ъ": "", "ы": "y", "ь": "", "э": "e", "ю": "yu", "я": "ya",
}

HEADER_RE = re.compile(
    r"^(\d+)\t([^\t]+?)\t+(\d+)\t+(.+)$"
)
# Fallback when tabs are messy: "N\tName\tNUM\t...file.jpg"
HEADER_LOOSE_RE = re.compile(
    r"^(\d+)\s+(.+?)\s+(\d+)\s+((?:.+\.(?:jpg|jpeg|png|webp))"
    r"(?:\s*\([^)]+\.(?:jpg|jpeg|png|webp)\))?)\s*$",
    re.IGNORECASE,
)

FIELD_KEYS = (
    "Выпуск", "Стихия", "Стоимость", "Тип", "Класс", "Свойства",
    "Иконки", "Текст", "Худ. текст", "Художник",
)

ATTACK_RE = re.compile(
    r"(?P<atype>(?:Простой|Выстрел|Метание|Разряд|Магический)\s+удар|"
    r"Простой удар|Выстрел|Метание|Разряд)\s+"
    r"(?P<dice>\d+-\d+-\d+)",
    re.IGNORECASE,
)
DICE_ONLY_RE = re.compile(r"\b(\d+-\d+-\d+)\b")
ICON_TOKEN_RE = re.compile(r"([a-zA-Zа-яА-ЯёЁ]+)(?::(\d+))?")


@dataclass
class CardItem:
    name: str
    card_no: int
    file_name: str
    image_path: str
    set_number: int
    year: int
    set_name: str
    rarity: str
    element: str
    cost: Optional[int]
    cost_tier: Optional[str]
    type_main: str
    type_sub: Optional[str]
    class_list: list[str]
    hp: Optional[int]
    move: Optional[int]
    attack_dice: Optional[str]
    attack_type: Optional[str]
    icons: list[dict[str, Any]]
    unique: bool
    flying: bool
    companion: bool
    symbiont: bool
    parasite: bool
    artist_display: str
    artist_id: str
    card_text: str
    flavor_text: str
    foil: bool
    noir: bool
    market_rub: int
    price: int
    stock: int
    source_info: str
    list_index: int = 0
    idempotency_key: str = ""

    def __post_init__(self) -> None:
        if not self.idempotency_key:
            self.idempotency_key = (
                f"{self.set_number}:{self.card_no}:"
                f"{'foil' if self.foil else 'plain'}:"
                f"{'noir' if self.noir else 'color'}:"
                f"{self.artist_id or 'unknown'}"
            )

    def attributes(self) -> dict[str, Any]:
        return {
            "set_number": self.set_number,
            "year": self.year,
            "set_name": self.set_name,
            "card_no": self.card_no,
            "rarity": self.rarity,
            "element": self.element,
            "cost": self.cost,
            "cost_tier": self.cost_tier,
            "type_main": self.type_main,
            "type_sub": self.type_sub,
            "class": self.class_list,
            "hp": self.hp,
            "move": self.move,
            "attack_dice": self.attack_dice,
            "attack_type": self.attack_type,
            "icons": self.icons,
            "unique": self.unique,
            "flying": self.flying,
            "companion": self.companion,
            "symbiont": self.symbiont,
            "parasite": self.parasite,
            "artist": self.artist_display,
            "artist_id": self.artist_id,
            "artist_display": self.artist_display,
            "card_text": self.card_text,
            "flavor_text": self.flavor_text,
            "foil": self.foil,
            "noir": self.noir,
            "file_name": self.file_name,
            "market_rub": self.market_rub,
            "idempotency_key": self.idempotency_key,
            "list_index": self.list_index,
        }

    def payload(self, image_url: Optional[str] = None) -> dict[str, Any]:
        images = [image_url] if image_url else (
            [f"(local){self.image_path}"] if self.image_path else []
        )
        return {
            "type": "карточка",
            "name": self.name,
            "description": self.card_text or self.flavor_text or "",
            "price": self.price,
            "stock": self.stock,
            "images": images,
            "attributes": self.attributes(),
        }


def load_dotenv(path: Path) -> None:
    if not path.is_file():
        return
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, val = line.partition("=")
        os.environ.setdefault(key.strip(), val.strip().strip("'\""))


def getenv(*keys: str, default: str = "") -> str:
    for k in keys:
        v = os.environ.get(k, "").strip()
        if v:
            return v
    return default


def slugify_artist(name: str) -> str:
    s = name.strip()
    if not s or s == "-":
        return "unknown"
    # Keep latin names like "Green Feline", "A. Kashen-Bazhenova", "SadTun"
    out: list[str] = []
    for ch in s.lower():
        if ch in _CYR:
            out.append(_CYR[ch])
        elif "a" <= ch <= "z" or "0" <= ch <= "9":
            out.append(ch)
        elif ch in " ._-":
            out.append("_")
        else:
            # strip combining marks from latinized leftovers
            n = unicodedata.normalize("NFKD", ch)
            for c in n:
                if "a" <= c.lower() <= "z":
                    out.append(c.lower())
    slug = re.sub(r"_+", "_", "".join(out)).strip("_")
    return slug or "unknown"


def normalize_name(name: str) -> str:
    n = name.strip()
    n = re.sub(r"\s*\(фойл\)\s*", " ", n, flags=re.IGNORECASE)
    n = re.sub(r"\s*\(нуар[^)]*\)\s*", " ", n, flags=re.IGNORECASE)
    n = re.sub(r"\s+", " ", n).strip().lower()
    return n


def price_for(rarity: str, foil: bool) -> tuple[int, int]:
    rub = RARITY_RUB.get(rarity, 10)
    if foil:
        rub *= 2
    return rub, rub * TICKETS_PER_RUB


def parse_icons(raw: str) -> list[dict[str, Any]]:
    raw = (raw or "").strip()
    if not raw or raw == "-":
        return []
    icons: list[dict[str, Any]] = []
    for m in ICON_TOKEN_RE.finditer(raw):
        t = m.group(1).lower()
        if t in ("и", "and"):
            continue
        val = m.group(2)
        icons.append({"type": t, "value": int(val) if val else None})
    return icons


def parse_properties(
    raw: str,
    type_main: str,
    type_line: str,
) -> dict[str, Any]:
    """Parse Свойства into hp/move/attack + flags."""
    result: dict[str, Any] = {
        "hp": None,
        "move": None,
        "attack_dice": None,
        "attack_type": None,
        "flying": "летающ" in (type_line or "").lower() or "летающ" in (raw or "").lower(),
        "companion": "компаньон" in (type_line or "").lower() or "компаньон" in (raw or "").lower(),
        "symbiont": "симбионт" in (type_line or "").lower() or "симбионт" in (raw or "").lower(),
        "parasite": "паразит" in (type_line or "").lower() or "паразит" in (raw or "").lower(),
        "unique": "уникальн" in (type_line or "").lower() or "уникальн" in (raw or "").lower(),
    }
    s = (raw or "").strip()
    if not s or s == "-":
        return result

    # Artifact / land: often "9 Артефакт" or just a number + word
    lower = s.lower()
    if type_main in ("artifact", "land") or "артефакт" in lower or "местность" in lower:
        nums = re.findall(r"\d+", s)
        if nums:
            result["hp"] = int(nums[0])
        return result

    atype = None
    dice = None
    m = ATTACK_RE.search(s)
    if m:
        atype = re.sub(r"\s+", " ", m.group("atype")).strip()
        dice = m.group("dice")
        before = s[: m.start()].strip()
    else:
        dm = DICE_ONLY_RE.search(s)
        if dm:
            dice = dm.group(1)
            before = s[: dm.start()].strip()
            # guess attack type words before dice
            words = before.split()
            # strip trailing tokens that are keywords
            atype = None
        else:
            before = s

    result["attack_dice"] = dice
    result["attack_type"] = atype

    # Strip known keywords from before-attack segment, collect numbers
    keywords = (
        "летающее", "летающий", "компаньон", "симбионт", "паразит",
        "уникальная", "уникальный", "бестелесный", "бестелесная",
        "наземный", "наземная",
    )
    tokens = before.split()
    nums: list[int] = []
    for tok in tokens:
        tl = tok.lower().rstrip(".,;")
        if tl in keywords or any(tl.startswith(k[:5]) for k in keywords if len(k) > 4):
            continue
        if re.fullmatch(r"\d+", tok):
            nums.append(int(tok))
    if nums:
        result["hp"] = nums[0]
        if len(nums) >= 2 and not result["flying"] and not result["companion"]:
            result["move"] = nums[1]
        elif len(nums) >= 2 and result["companion"]:
            # companion: usually hp only before keyword; second num rare
            result["move"] = None
        elif len(nums) >= 2 and result["flying"]:
            # flying often has no move digit
            result["move"] = None
    return result


def parse_type(raw: str) -> tuple[str, Optional[str]]:
    raw = (raw or "").strip()
    if not raw:
        return "creature", None
    main_part, _, sub = raw.partition("–")
    if not sub:
        main_part, _, sub = raw.partition("-")
    main = main_part.strip().lower()
    type_main = TYPE_MAIN_RU.get(main, main if main in ("creature", "artifact", "land") else "creature")
    # map russian main
    for ru, en in TYPE_MAIN_RU.items():
        if main.startswith(ru):
            type_main = en
            break
    type_sub = sub.strip() or None
    return type_main, type_sub


def parse_выпуск(raw: str) -> tuple[str, str, Optional[int]]:
    """Return rarity, set_name, card_no_from_line."""
    parts = (raw or "").split()
    rarity = "common"
    if parts:
        r0 = parts[0].lower()
        rarity = RARITY_RU.get(r0, rarity)
        rest = parts[1:]
    else:
        rest = []
    card_no = None
    if rest and rest[-1].isdigit():
        card_no = int(rest[-1])
        rest = rest[:-1]
    set_name = " ".join(rest).strip()
    return rarity, set_name, card_no


def parse_cost(raw: str) -> tuple[Optional[int], Optional[str]]:
    parts = (raw or "").split()
    cost = int(parts[0]) if parts and parts[0].isdigit() else None
    tier = None
    if len(parts) >= 2:
        tier = COST_TIER_RU.get(parts[1].lower(), parts[1].lower())
    return cost, tier


def extract_filenames(file_field: str) -> list[str]:
    field = file_field.strip()
    files: list[str] = []
    # primary + optional (secondary) — allow spaces in names
    m = re.match(
        r"^(.+?\.(?:jpg|jpeg|png|webp))\s*(?:\(([^)]+\.(?:jpg|jpeg|png|webp))\))?\s*$",
        field,
        re.IGNORECASE,
    )
    if m:
        files.append(m.group(1).strip())
        if m.group(2):
            files.append(m.group(2).strip())
        return files
    files = re.findall(
        r"[^\s()][^()]*?\.(?:jpg|jpeg|png|webp)",
        field,
        flags=re.IGNORECASE,
    )
    return [f.strip() for f in files]


def set_meta_from_path(info_path: Path) -> tuple[int, int]:
    """Infer set_number and year from path segments like 4th_2024 / 8th_2026."""
    text = str(info_path)
    m = re.search(r"/(\d+)(?:st|nd|rd|th)_(\d{4})/", text)
    if m:
        return int(m.group(1)), int(m.group(2))
    m2 = re.search(r"/(\d{4})/", text)
    year = int(m2.group(1)) if m2 else 2026
    return 0, year


def split_card_blocks(text: str) -> list[str]:
    """Split info file into blocks starting at header lines."""
    lines = text.splitlines()
    starts: list[int] = []
    for i, line in enumerate(lines):
        if i == 0 and line.startswith("№"):
            continue
        if HEADER_RE.match(line) or HEADER_LOOSE_RE.match(line):
            starts.append(i)
            continue
        # also accept "N\tName\tNUM\t...jpg"
        if re.match(r"^\d+\t", line) and re.search(r"\.(?:jpg|jpeg|png|webp)", line, re.I):
            starts.append(i)
    blocks: list[str] = []
    for i, start in enumerate(starts):
        end = starts[i + 1] if i + 1 < len(starts) else len(lines)
        blocks.append("\n".join(lines[start:end]))
    return blocks


def parse_header_line(line: str) -> Optional[tuple[int, str, int, str]]:
    m = HEADER_RE.match(line)
    if m:
        return int(m.group(1)), m.group(2).strip(), int(m.group(3)), m.group(4).strip()
    m = HEADER_LOOSE_RE.match(line.strip())
    if m:
        return int(m.group(1)), m.group(2).strip(), int(m.group(3)), m.group(4).strip()
    # tab-heavy: split on tabs and filter empties
    parts = [p for p in line.split("\t") if p.strip()]
    if len(parts) >= 4 and parts[0].isdigit() and parts[2].isdigit():
        return int(parts[0]), parts[1].strip(), int(parts[2]), parts[3].strip()
    return None


def parse_fields(block_lines: list[str]) -> dict[str, str]:
    fields: dict[str, str] = {}
    current: Optional[str] = None
    buf: list[str] = []
    after_extra_file = False

    def flush() -> None:
        nonlocal current, buf
        if current is not None:
            fields[current] = "\n".join(buf).strip()
        current, buf = None, []

    for line in block_lines:
        matched = False
        for key in FIELD_KEYS:
            if line.startswith(key + "\t") or line.startswith(key + " "):
                flush()
                rest = line[len(key):].lstrip("\t ")
                # Second+ Художник after an extra filename → secondary artist
                if key == "Художник" and ("Художник" in fields or after_extra_file):
                    fields["_secondary_artist"] = rest.strip()
                    current = None
                    buf = []
                    after_extra_file = False
                    matched = True
                    break
                current = key
                buf = [rest] if rest else []
                matched = True
                break
        if matched:
            continue
        # "Художник Green Feline" (space, no tab) after extra file
        ls = line.strip()
        if after_extra_file and re.match(r"^Художник\s+", ls):
            flush()
            fields["_secondary_artist"] = re.sub(r"^Художник\s+", "", ls).strip()
            after_extra_file = False
            continue
        # secondary artist block: bare filename then Художник
        if re.match(r"^[^\s/]+\.(?:jpg|jpeg|png|webp)\s*$", ls, re.I):
            flush()
            fields.setdefault("_extra_files", "")
            fields["_extra_files"] = (fields.get("_extra_files", "") + "\n" + ls).strip()
            after_extra_file = True
            continue
        if current is not None:
            # continuation of multi-line Текст etc.
            if line.strip() == "" and current not in ("Текст", "Худ. текст"):
                flush()
            else:
                buf.append(line)
    flush()
    return fields


def _norm_fname(name: str) -> str:
    """Collapse spaces for fuzzy image match (e.g. 'zh nets' ↔ 'zhnets')."""
    return re.sub(r"\s+", "", name).lower()


def resolve_image(folder: Path, file_name: str) -> Optional[Path]:
    p = folder / file_name
    if p.is_file():
        return p
    # literal with spaces preserved from info (disk may have spaces)
    lower = file_name.lower()
    target = _norm_fname(file_name)
    target_stem = _norm_fname(Path(file_name).stem)
    best: Optional[Path] = None
    for cand in folder.iterdir():
        if not cand.is_file():
            continue
        if cand.name.lower() == lower:
            return cand
        cn = _norm_fname(cand.name)
        if cn == target:
            return cand
        # 002_zhongler_198_common.jpg ↔ 002_zhongler_198_common.png.jpg
        cs = _norm_fname(cand.stem)
        # strip extra .png from stem of *.png.jpg
        cs2 = cs.removesuffix(".png").removesuffix(".jpeg")
        if cs == target_stem or cs2 == target_stem:
            best = cand
    return best


def parse_info_file(info_path: Path) -> list[CardItem]:
    set_number, year = set_meta_from_path(info_path)
    text = info_path.read_text(encoding="utf-8")
    folder = info_path.parent
    items: list[CardItem] = []

    for block in split_card_blocks(text):
        lines = block.splitlines()
        if not lines:
            continue
        header = parse_header_line(lines[0])
        if not header:
            continue
        list_index, name_raw, card_no, file_field = header
        fields = parse_fields(lines[1:])

        rarity, set_name, _ = parse_выпуск(fields.get("Выпуск", ""))
        element_raw = fields.get("Стихия", "").strip().lower()
        element = ELEMENT_RU.get(element_raw, element_raw or "neutral")
        cost, cost_tier = parse_cost(fields.get("Стоимость", ""))
        type_main, type_sub = parse_type(fields.get("Тип", ""))
        class_raw = fields.get("Класс", "").strip()
        class_list = [c.strip() for c in class_raw.split(",") if c.strip()] if class_raw and class_raw != "-" else []
        props = parse_properties(fields.get("Свойства", ""), type_main, fields.get("Тип", ""))
        icons = parse_icons(fields.get("Иконки", ""))
        card_text = fields.get("Текст", "").strip()
        flavor = fields.get("Худ. текст", "").strip()
        if flavor == "-":
            flavor = ""
        artist_primary = fields.get("Художник", "").strip()
        secondary_artist = fields.get("_secondary_artist", "").strip()

        files = extract_filenames(file_field)
        extra = fields.get("_extra_files", "")
        for ef in extra.splitlines():
            ef = ef.strip()
            if ef and ef not in files:
                files.append(ef)

        artists = [artist_primary]
        if secondary_artist:
            artists.append(secondary_artist)
        while len(artists) < len(files):
            artists.append(artists[-1] if artists else "unknown")

        foil_name = "фойл" in name_raw.lower() or "foil" in file_field.lower()
        noir_name = "нуар" in name_raw.lower() or "noir" in file_field.lower()
        display_name = re.sub(r"\s*\(фойл\)\s*", "", name_raw, flags=re.I)
        display_name = re.sub(r"\s*\(нуар[^)]*\)\s*", "", display_name, flags=re.I).strip()
        display_name = re.sub(r"\s+", " ", display_name)

        for fi, fname in enumerate(files):
            artist_display = artists[fi] if fi < len(artists) else artist_primary
            artist_id = slugify_artist(artist_display)
            foil = foil_name or ("foil" in fname.lower())
            noir = noir_name or ("noir" in fname.lower())
            market_rub, price = price_for(rarity, foil)
            img = resolve_image(folder, fname)
            items.append(
                CardItem(
                    name=display_name,
                    card_no=card_no,
                    file_name=fname,
                    image_path=str(img) if img else "",
                    set_number=set_number,
                    year=year,
                    set_name=set_name,
                    rarity=rarity,
                    element=element,
                    cost=cost,
                    cost_tier=cost_tier,
                    type_main=type_main,
                    type_sub=type_sub,
                    class_list=class_list,
                    hp=props["hp"],
                    move=props["move"],
                    attack_dice=props["attack_dice"],
                    attack_type=props["attack_type"],
                    icons=icons,
                    unique=props["unique"],
                    flying=props["flying"],
                    companion=props["companion"],
                    symbiont=props["symbiont"],
                    parasite=props["parasite"],
                    artist_display=artist_display,
                    artist_id=artist_id,
                    card_text=card_text,
                    flavor_text=flavor,
                    foil=foil,
                    noir=noir,
                    market_rub=market_rub,
                    price=price,
                    stock=1,
                    source_info=str(info_path.relative_to(ROOT)),
                    list_index=list_index,
                )
            )
    return items


# README typos / spelling variants → card info names (normalized)
STOCK_ALIASES = {
    "бешенный карлик": "бешеный карлик",
    "гарканн": "гаркаин",
    "жнец лотоса": "жнец лотоса",
}


def parse_stock_readme(path: Path) -> dict[str, int]:
    """Map normalized card name (+ optional foil) → stock count.

    Keys: 'name', 'name|foil', 'name|noir'. Multi-copy (х2) without Nшт → 1 each.
    """
    stock: dict[str, int] = {}
    if not path.is_file():
        return stock
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if (
            not line
            or line.endswith(":")
            or line.startswith("CARDS")
            or re.match(r"^[4-8]th\b", line)
        ):
            continue
        if re.match(r"^(swamps|neutrals|mountains|woods|steppes|darkness):?\s*$", line):
            continue
        m = re.search(r"(\d+)\s*шт", line)
        count = int(m.group(1)) if m else None
        name_part = line[: m.start()].strip() if m else line
        # drop trailing notes after em-dash / hyphen comment
        name_part = re.split(r"\s+[—\-]\s+", name_part, maxsplit=1)[0].strip()

        foil = bool(re.search(r"\(фойл\)", name_part, re.I)) or (
            "фойл" in name_part.lower() and ("фойлов" in name_part.lower() or "обе фойл" in name_part.lower())
        )
        noir = "нуар" in name_part.lower()
        if count is None and re.search(r"\(х\s*2", name_part, re.I):
            count = 1
            foil = foil or "фойл" in name_part.lower()
        if count is None:
            continue

        # Base name = text before first '(' (handles nested / long notes)
        base = name_part.split("(", 1)[0].strip()
        # also strip simple trailing parentheticals if any leftover
        base = re.sub(r"\s+", " ", base).strip().lower()
        if not base:
            continue
        base = STOCK_ALIASES.get(base, base)

        if foil:
            stock[f"{base}|foil"] = count
        if noir:
            stock[f"{base}|noir"] = count
        stock.setdefault(base, count)
        if foil or noir:
            stock[base] = count
    return stock


def apply_stock(items: list[CardItem], stock_map: dict[str, int]) -> list[str]:
    unresolved: list[str] = []
    for it in items:
        key = normalize_name(it.name)
        key = STOCK_ALIASES.get(key, key)
        candidates = []
        if it.foil:
            candidates.append(f"{key}|foil")
        if it.noir:
            candidates.append(f"{key}|noir")
        candidates.append(key)
        found = None
        for c in candidates:
            if c in stock_map:
                found = stock_map[c]
                break
        if found is not None:
            it.stock = found
        else:
            unresolved.append(f"{it.name} (set {it.set_number}, {it.file_name})")
            it.stock = 1
    return unresolved


def discover_all() -> list[CardItem]:
    items: list[CardItem] = []
    for info in sorted(CARDS_ROOT.rglob("*_cards_info.md")):
        items.extend(parse_info_file(info))
    stock_map = parse_stock_readme(STOCK_README)
    apply_stock(items, stock_map)
    return items


def filter_smoke(items: list[CardItem]) -> list[CardItem]:
    selected: list[CardItem] = []
    for name_key, set_no in SMOKE_ORDER:
        matches = [
            it for it in items
            if normalize_name(it.name) == name_key and it.set_number == set_no
        ]
        if not matches:
            print(f"WARN: smoke card not found: {name_key} set {set_no}", file=sys.stderr)
            continue
        # mormolika: both artists; others: first match (prefer foil if name implies)
        if name_key == "мормолика":
            selected.extend(matches)
        elif name_key == "тхакрай":
            foil = [m for m in matches if m.foil]
            selected.append((foil or matches)[0])
        else:
            selected.append(matches[0])
    return selected


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
        data = json.dumps(body).encode("utf-8")
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


def http_multipart_file(url: str, token: str, file_path: Path) -> dict:
    import uuid

    boundary = f"----SeedBoundary{uuid.uuid4().hex}"
    filename = file_path.name
    content = file_path.read_bytes()
    ctype = "image/jpeg"
    if filename.lower().endswith(".png"):
        ctype = "image/png"
    elif filename.lower().endswith(".webp"):
        ctype = "image/webp"

    parts = []
    parts.append(f"--{boundary}\r\n".encode())
    parts.append(
        f'Content-Disposition: form-data; name="file"; filename="{filename}"\r\n'
        f"Content-Type: {ctype}\r\n\r\n".encode()
    )
    parts.append(content)
    parts.append(f"\r\n--{boundary}--\r\n".encode())
    body = b"".join(parts)

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": f"multipart/form-data; boundary={boundary}",
        "Accept": "application/json",
    }
    req = Request(url, data=body, headers=headers, method="POST")
    try:
        with urlopen(req, timeout=120) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except HTTPError as e:
        err_body = e.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"upload HTTP {e.code}: {err_body}") from e


def login(base: str) -> str:
    token = getenv("SEED_TOKEN")
    if token:
        return token
    email = getenv(
        "SEED_EMAIL", "SEED_ADMIN_EMAIL", "SEED_AUTHOR_EMAIL",
        default="admin@eventhorizon.local",
    )
    password = getenv(
        "SEED_PASSWORD", "SEED_ADMIN_PASSWORD", "SEED_AUTHOR_PASSWORD",
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
    return tok


def fetch_existing_keys(base: str, token: str) -> dict[str, dict]:
    """Map idempotency_key → item dict for type=карточка."""
    existing: dict[str, dict] = {}
    offset = 0
    limit = 100
    while True:
        q = urlencode({"type": "карточка", "limit": limit, "offset": offset})
        resp = http_json("GET", f"{base}/api/inventory/items?{q}", token=token)
        batch = resp.get("items") or []
        total = int(resp.get("total") or 0)
        for it in batch:
            attrs = it.get("attributes") or {}
            key = attrs.get("idempotency_key")
            if key:
                existing[key] = it
        offset += len(batch)
        if not batch or offset >= total:
            break
    return existing


def print_dry_run(items: list[CardItem]) -> None:
    print(f"=== DRY-RUN: {len(items)} item(s) ===\n")
    for i, it in enumerate(items, 1):
        missing = "" if it.image_path else " [MISSING IMAGE]"
        print(f"--- {i}. {it.name}{missing} ---")
        print(json.dumps(it.payload(), ensure_ascii=False, indent=2))
        print()


def run_seed(items: list[CardItem], base: str, dry_run: bool, update: bool) -> None:
    if dry_run:
        print_dry_run(items)
        return

    token = login(base)
    existing = fetch_existing_keys(base, token)
    created = skipped = updated = errors = 0

    for it in items:
        key = it.idempotency_key
        if key in existing and not update:
            print(f"SKIP  {it.name} [{key}] → {existing[key].get('id')}")
            skipped += 1
            continue
        if not it.image_path or not Path(it.image_path).is_file():
            print(f"ERROR missing image for {it.name} ({it.file_name})", file=sys.stderr)
            errors += 1
            continue
        try:
            up = http_multipart_file(f"{base}/api/uploads", token, Path(it.image_path))
            url = up.get("url")
            if not url:
                raise RuntimeError(f"no url in upload response: {up}")
            body = it.payload(url)
            if key in existing and update:
                item_id = existing[key]["id"]
                resp = http_json(
                    "PUT",
                    f"{base}/api/inventory/items/{item_id}",
                    token=token,
                    body=body,
                )
                print(f"UPDATE {it.name} [{key}] → {item_id}")
                updated += 1
            else:
                resp = http_json(
                    "POST",
                    f"{base}/api/inventory/items",
                    token=token,
                    body=body,
                )
                item = resp.get("item") or resp
                print(f"CREATE {it.name} [{key}] → {item.get('id')} price={it.price} stock={it.stock}")
                created += 1
                if item.get("id"):
                    existing[key] = item
        except Exception as e:
            print(f"ERROR {it.name}: {e}", file=sys.stderr)
            errors += 1

    print(
        f"\nDone: created={created} updated={updated} skipped={skipped} errors={errors}"
    )


def main() -> int:
    load_dotenv(ROOT / "scripts" / ".env.seed.admin")
    load_dotenv(ROOT / "scripts" / ".env.seed.author")

    ap = argparse.ArgumentParser(description="Seed Berserk cards into inventory")
    ap.add_argument("--dry-run", action="store_true", help="Print payloads, no writes")
    ap.add_argument("--smoke", action="store_true", help="Only the 5 smoke cards (Мормолика×2)")
    ap.add_argument("--limit", type=int, default=0, help="Max items after filter")
    ap.add_argument("--update", action="store_true", help="PUT existing idempotency keys")
    ap.add_argument(
        "--gateway",
        default=getenv("SEED_GATEWAY_URL", "GATEWAY_URL", default="http://localhost:8079"),
        help="Gateway base URL",
    )
    args = ap.parse_args()

    if not CARDS_ROOT.is_dir():
        print(f"Cards root missing: {CARDS_ROOT}", file=sys.stderr)
        return 1

    items = discover_all()
    print(f"Parsed {len(items)} card items from {CARDS_ROOT}", file=sys.stderr)

    missing_img = [it for it in items if not it.image_path]
    if missing_img:
        print(f"WARN: {len(missing_img)} items missing local images", file=sys.stderr)

    stock_map = parse_stock_readme(STOCK_README)
    unresolved = apply_stock(items, stock_map)
    if unresolved:
        print(f"WARN: {len(unresolved)} items used default stock=1 (no README match)", file=sys.stderr)

    if args.smoke:
        items = filter_smoke(items)
        print(f"Smoke filter → {len(items)} items", file=sys.stderr)

    if args.limit and args.limit > 0:
        items = items[: args.limit]
        print(f"Limit → {len(items)} items", file=sys.stderr)

    run_seed(items, args.gateway.rstrip("/"), dry_run=args.dry_run, update=args.update)
    return 0


if __name__ == "__main__":
    sys.exit(main())
