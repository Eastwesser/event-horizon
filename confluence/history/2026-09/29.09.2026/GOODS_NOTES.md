# Goods — Берсерк / ККИ cards (from voice notes 29.09.2026)

## Product type
Inventory `type = "карточка"` (alongside брелок / картина / фенечка).

## Pricing practice (not live marketplace)
- Rough map: **1 ₽ ≈ 2000 билетиков** → e.g. ~10₽ card → **20 000** tickets.
- **Foil ×2**.
- Stock is per-copy (often 1–3). No FOMO: sold-out is sold-out; duplicates only by agreed exception.

## Attributes (stored in `attributes` JSON)
| Key | Example | Notes |
|-----|---------|--------|
| `element` | лес / горы / степи / тьма / болото / нейтрал | 6 «стихий» for shop filters |
| `rarity` | common / uncommon / rare / ultra | green→gold ladder |
| `set` | 7 | set number |
| `year` | 2006 | release year |
| `artist` | ark_cj | for author filter later |
| `foil` | true/false | |
| `card_no` | — | number in series |
| `market_rub` | 10 | human estimate |

## Sort / seed later
Prefer: set → artist → year → name (alpha). Combat stats ignored for shop MVP.

## Status
Create modal + filter chip for `карточка` shipped. Bulk photo→catalog seed is manual follow-up (~150–300 cards).
