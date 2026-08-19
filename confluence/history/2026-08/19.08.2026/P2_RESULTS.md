P2 is in. The three leftover items from the 14.08 list are implemented and the unit smokes pass.

Unit smokes

Balancer: least-conn picks the backend with the fewest active connections.
Fulfillment: HandlePurchasePaid with a 1ms delay publishes PurchaseFulfilled through a stub Kafka producer.
Notification: ignored topic, paid event with empty Telegram token (no HTTP), bad JSON.
The nested services/balancer/internal/balancer/go.mod is gone so that package is a normal part of the balancer module (that was blocking go test).

Shop outbox A purchase now writes shop.purchased into outbox in the same Postgres transaction as stock/inventory/purchases. A worker publishes to NATS, same pattern as billing/inventory. Compensation shop.purchase.failed is still a direct publish. Kafka purchase.paid is unchanged (after commit).

slog log.Printf in payment, authors, billing, inventory, and shop outbox workers, plus gateway WebSocket connect/disconnect, now uses log/slog.

The shop image was not rebuilt or pushed. For the cluster: rebuild shop, then make migrate-shop (or let shop apply migrations on start). I have not committed; say if you want a commit and/or a shop image push.