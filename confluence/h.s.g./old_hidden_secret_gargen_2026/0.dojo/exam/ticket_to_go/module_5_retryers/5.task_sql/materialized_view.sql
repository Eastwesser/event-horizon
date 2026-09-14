-- Module 5 / SQL: кэш тяжёлого отчёта через Materialized View

CREATE MATERIALIZED VIEW expensive_report AS
SELECT
    user_id,
    COUNT(*) AS total_orders,
    SUM(amount) AS total_spent
FROM orders
GROUP BY user_id;

-- CONCURRENTLY требует уникальный индекс на MV
CREATE UNIQUE INDEX ON expensive_report (user_id);
REFRESH MATERIALIZED VIEW CONCURRENTLY expensive_report;

-- На собесе: MV = кэш в БД; Redis — кэш у приложения.
-- Refresh — отдельный job; stale-while-revalidate как продуктовый компромисс.
