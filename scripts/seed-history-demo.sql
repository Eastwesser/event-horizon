-- Demo history events for admin (eventhorizon_history).
-- Replace USER_ID if needed. Default = seeded admin UUID.
-- Usage (example):
--   psql "$HISTORY_DSN" -f scripts/seed-history-demo.sql

BEGIN;

DO $$
DECLARE
  uid uuid := '1502a3fa-0e64-4873-a329-3d8fa1d5204d';
BEGIN
  INSERT INTO events (id, user_id, event_type, payload, created_at)
  VALUES
    (gen_random_uuid(), uid, 'user.registered',
     jsonb_build_object('user_id', uid::text, 'email', 'admin@eventhorizon.local'),
     NOW() - INTERVAL '2 days'),
    (gen_random_uuid(), uid, 'score.updated',
     jsonb_build_object('user_id', uid::text, 'game_id', 'flappy', 'score', 50, 'level', 1),
     NOW() - INTERVAL '1 day'),
    (gen_random_uuid(), uid, 'shop.purchased',
     jsonb_build_object('user_id', uid::text, 'item_name', 'Космический брелок', 'price', 100000),
     NOW() - INTERVAL '12 hours'),
    (gen_random_uuid(), uid, 'payment.completed',
     jsonb_build_object('user_id', uid::text, 'plan', 'present', 'status', 'active'),
     NOW() - INTERVAL '6 hours'),
    (gen_random_uuid(), uid, 'author.upserted',
     jsonb_build_object('user_id', uid::text, 'display_name', 'Event Horizon'),
     NOW() - INTERVAL '3 hours')
  ON CONFLICT DO NOTHING;
END $$;

COMMIT;
