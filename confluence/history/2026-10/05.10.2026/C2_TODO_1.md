C1 verified locally. Push 7110d1b, then start C2.

Push:
  - 7110d1b (Author application C1 + register role fix)
Report new HEAD.

================================================================
WAVE 3 C2 — Admin approval (M)
================================================================

Plan first, no code.

Scope:
  - /admin → new tab «Заявки» (next to Пользователи /
    Инвентарь / Аналитика).
  - Table: applicant (email + display_name), submitted_at,
    status.
  - Filter: pending | approved | rejected. Default: pending.
  - Actions:
      - Approve → POST /api/authors/applications/:id/approve
        → status=approved + Auth.UpdateRole(user_id, author)
        → optionally welcome notify (deferred; just note).
      - Reject (with optional note) → status=rejected +
        reviewer_note.
  - RBAC: admin only.
  - Audit: reviewed_at, reviewed_by (admin user id).

Questions to answer before code:
  1. Where to call Auth.UpdateRole? From gateway, or
     from Authors service via gRPC? Recommended: gateway
     orchestrates — it already calls Auth for other
     admin ops. Confirm.
  2. Approval is not idempotent by nature — should a second
     approve return 400 "already reviewed"? Recommended
     yes.
  3. After approve — does the user's next request to Auth
     pick up the new role (JWT refresh)? If not, note it
     as a UX gap.
  4. Admin UI tab placeholder exists? If not, plan a
     minimal table + filter.

Files expected:
  - Gateway: POST approve/reject endpoints, list applications
    endpoint (admin-only).
  - Authors: proto ApproveApplication / RejectApplication /
    ListApplications.
  - FE: /admin tab «Заявки» + table + actions.
  - Migration: none (columns already in C1).

Rules:
  - C2-only PR. Do NOT touch C1, C3/C4, other waves,
    hero, nav, disk, games, seed, refund.
  - tsc + build clean.
  - Smoke:
      - Submit application as user.
      - Admin sees it in the table.
      - Approve → user becomes author.
      - Re-login as that user → /register-author redirects
        (already author).
      - Reject on another → status=rejected, note saved.
      - Double approve → 400.
  - Screenshots: admin tab, approve flow.
  - No push until I OK.


  ALSO!

  ✔ Container event-horizon-grafana          Running                                                                                 0.0s
 ✔ Container deployments-notification-1     Running                                                                                 0.0s
 ✔ Container deployments-fulfillment-1      Running                                                                                 0.0s
 ✔ Container nats-exporter                  Running                                                                                 0.0s
 ✔ Container deployments-analytics-1        Running                                                                                 0.0s
 ✘ Container deployments-billing-1          Error dependency billing failed to start                                               33.5s
 ✔ Container event-horizon-jaeger           Running                                                                                 0.0s
 ✔ Container event-horizon-prometheus       Running                                                                                 0.0s
 ✔ Container event-horizon-postgres-billing Started                                                                                15.5s
 ✔ Container event-horizon-redis-inventory  Started                                                                                25.2s
 ✔ Container event-horizon-postgres         Healthy                                                                                29.7s
 ... 30 more                                                                                                                            
dependency failed to start: container deployments-billing-1 is unhealthy
make: *** [Makefile:150: deploy] Ошибка 1
[denismatveev@c0der event_horizon]$ 

Then I do "make deploy" once again and everything starts OK.  Look:

 ✔ Container event-horizon-postgres           Healthy                                                                               0.6s
 ✔ Container event-horizon-postgres-inventory Running                                                                               0.0s
 ✔ Container event-horizon-postgres-profile   Healthy                                                                               0.6s
 ✔ Container event-horizon-redis-exporter     Running                                                                               0.0s
 ✔ Container event-horizon-redis              Healthy                                                                               0.6s
 ✔ Container event-horizon-grafana            Running                                                                               0.0s
 ✔ Container event-horizon-redis-shop         Running                                                                               0.0s
 ✔ Container deployments-history-1            Running                                                                               0.0s
 ✔ Container event-horizon-nats-1             Healthy                                                                               0.6s
 ✔ Container deployments-gateway-1            Running                                                                               0.0s
 ✔ Container deployments-billing-1            Healthy                                                                               0.6s
 ✔ Container nats-exporter                    Running                                                                               0.0s
 ... 36 more                                                                                                                            
📦 Running migrations...
make migrate-all
make[1]: вход в каталог «/home/denismatveev/event_horizon»
cd services/auth && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable" up
2026/10/05 13:19:26 goose: no migrations to run. current version: 20260813000000
cd services/billing && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable" up
2026/10/05 13:19:26 goose: no migrations to run. current version: 20260920170000
cd services/game && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5461/eventhorizon_game?sslmode=disable" up
2026/10/05 13:19:26 goose: no migrations to run. current version: 20261004020000
cd services/leaderboard && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5463/eventhorizon_leaderboard?sslmode=disable" up
2026/10/05 13:19:26 goose: no migrations to run. current version: 20261004020000
cd services/profile && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5464/eventhorizon_profile?sslmode=disable" up
2026/10/05 13:19:26 goose: no migrations to run. current version: 20261004030000
cd services/shop && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5465/eventhorizon_shop?sslmode=disable" up
2026/10/05 13:19:27 goose: no migrations to run. current version: 20261003050000
cd services/inventory && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5466/eventhorizon_inventory?sslmode=disable" up
2026/10/05 13:19:27 goose: no migrations to run. current version: 20260813190000
✅ All migrations applied
make[1]: выход из каталога «/home/denismatveev/event_horizon»
✅ Thin stack ready
   🎯 Gateway: http://localhost:8079
   📊 Grafana: http://localhost:3000
   🔍 Jaeger: http://localhost:16686
   📈 Prometheus: http://localhost:9090
   Stronger box later: make deploy-heavy  (adds Kafka broker)
   Stop Kafka only: make stop-heavy
[denismatveev@c0der event_horizon]$ 
