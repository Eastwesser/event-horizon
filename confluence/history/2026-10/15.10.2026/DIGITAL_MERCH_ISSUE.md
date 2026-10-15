1. Как отличить цифровой товар от физического
Смотрю на ShopItem (в shopStore.ts):

ts
export interface ShopItem {
  id: string;
  name: string;
  description: string;
  price_tickets: number;
  icon_url: string;
  type: string;         // ← это либо category, либо type
  category: string;     // ← сюда попадает category из item
  game_id?: string;     // ← вот это!
  image_url?: string;
  images?: string[];
  available: boolean;
  owned: boolean;
  stock?: number | null;
  attributes?: Record<string, unknown>;
  created_at?: string;
}
Ключ — game_id. У скинов он заполнен (gears, flappy, hexagon, towers, memory, companion). У физического мерча его нет.

Значит, в ShopItemCard.tsx и ShopItemDetail.tsx можно сделать так:

tsx
const isDigital = Boolean(item.game_id);
И скрывать stockText для цифровых товаров.

Также stockLabel в src/lib/shopItemMap.ts — скинь его. Возможно, там уже есть логика, которую можно расширить.

И ещё: в ShopItemDetail.tsx блок с stockText выглядит так:

tsx
{stockText !== null && (
  <span className={outOfStock ? 'text-text-muted' : 'text-text-secondary'}>
    {stockText}
  </span>
)}
Меняем на:

tsx
{stockText !== null && !isDigital && (
  <span className={outOfStock ? 'text-text-muted' : 'text-text-secondary'}>
    {stockText}
  </span>
)}
Аналогично в ShopItemCard.tsx:

tsx
{stockText !== null && !isDigital && (
  <p className={...}>
    {stockText}
  </p>
)}
Но! Есть нюанс: если stock реально равен 0 (например, физический мерч закончился), то outOfStock = true, и кнопка «Купить» должна быть disabled. Для цифровых товаров stock = 999 или null, и outOfStock = false — всё ок.

UPDATE

У тебя в файле описан правильный подход, но одна деталь устарела:

Ты пишешь:

tsx
const isDigital = Boolean(item.game_id);
Но в ShopItemDetail.tsx item — это InventoryItem, у которого нет поля game_id. Оно лежит в item.attributes.game_id.

Правильно:

tsx
const attrs = (item.attributes || {}) as Record<string, unknown>;
const isDigital = Boolean(attrs.game_id);
А в ShopItemCard.tsx item — это ShopItem, у которого game_id есть и напрямую, и в attributes. Там оба варианта работают, но лучше единообразно:

tsx
const attrs = (item.attributes || {}) as Record<string, unknown>;
const isDigital = Boolean(attrs.game_id);
Допиши в файл:

В ShopItemDetail.tsx item — InventoryItem (нет game_id в корне). Брать attrs.game_id.
В ShopItemCard.tsx item — ShopItem, есть game_id в корне, но для единообразия брать attrs.game_id.

Допиши в DIGITAL_MERCH_ISSUE.md:

isDigital должен учитывать не только game_id, но и category/type (game_skin, profile_theme, theme, скин, тема). Цифровые: скины, темы, игровые предметы. Физические: карточки, мерч, брелоки, картины, фенечки.

Допиши в DIGITAL_MERCH_ISSUE.md:

Для цифровых товаров (скины, темы) не использовать images из инвентаря — там лежит обложка физической карточки. Всегда показывать fallback-иконку (иконка игры для скинов, sparkle для тем).