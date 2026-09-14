-- Module 4 / SQL: SLA / uptime за 24 часа

SELECT
    service_name,
    ROUND(COUNT(*) FILTER (WHERE status = 'up') * 100.0 / NULLIF(COUNT(*), 0), 3) AS uptime_percent,
    ROUND(AVG(response_time_ms)::numeric, 2) AS avg_response_time_ms,
    COUNT(*) FILTER (WHERE status <> 'up') AS failed_checks
FROM health_checks
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY service_name
ORDER BY uptime_percent ASC;

-- На собесе:
-- 99.9% uptime ≈ 43 мин даунтайма/месяц; 99.99% ≈ 4 мин.
-- liveness ≠ readiness; SLA считают по user-facing ошибкам, не только по /health.
