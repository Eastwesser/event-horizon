-- Wave 2 #2 — drop unpurchased placeholder брелоки (2026-10-04).
-- Keep: all карточка, all game_skin / profile_theme, purchased placeholders.
-- Report-only (do not delete): Блинный мерч (purchased).
--
-- Backups taken before apply:
--   ~/backup_20261004_011746_inventory.sql
--   ~/backup_20261004_011746_shop.sql
--
-- Applied against live cluster. Idempotent re-run is safe.

-- Inventory: soft-delete unpurchased placeholders
UPDATE inventory_items
SET deleted_at = COALESCE(deleted_at, NOW()),
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND id IN (
    '1590b92b-55cc-4cd2-baae-83424fd258c3', -- Автоматический брелок
    '6e4d8ba9-b0ad-439c-94b0-ec14c8312899', -- Автоматический брелок
    '39b556df-2079-45f3-b3b8-cce26aa4df13', -- Медный брелок
    '37c3136f-ea99-48ae-ad1d-596d8b6371cd'  -- Тестовый брелок
  );

-- Shop: hard-delete only if never purchased / never owned
DELETE FROM items
WHERE id IN (
    '1590b92b-55cc-4cd2-baae-83424fd258c3',
    '6e4d8ba9-b0ad-439c-94b0-ec14c8312899',
    '39b556df-2079-45f3-b3b8-cce26aa4df13',
    '37c3136f-ea99-48ae-ad1d-596d8b6371cd'
  )
  AND NOT EXISTS (SELECT 1 FROM inventory i WHERE i.item_id = items.id)
  AND NOT EXISTS (SELECT 1 FROM purchases p WHERE p.item_id = items.id);

-- KEPT (purchased — must stay for users):
--   c96419b1-f283-4fb5-b03f-d73baf6efd52 Автоматический брелок
--   4fb5c062-d212-4dff-bab1-5d346f76adff Брелок с медведем
--   ff137485-f356-4ada-9b00-bfb84546204a Ключница Дракон
--   5e3254d2-ed02-4584-af17-dcd3b4b0cfef Супер-пупер брелок
--   b5b332fd-146d-464c-b5f3-389a8fa48b83 Блинный мерч (report-only)
