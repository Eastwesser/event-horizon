-- Module 3 / SQL: error rate по часам (что смотреть, когда решать открывать CB / алертить).

SELECT
    DATE_TRUNC('hour', timestamp) AS hour,
    COUNT(*) FILTER (WHERE status >= 500) AS errors,
    COUNT(*) AS total,
    ROUND(
        (COUNT(*) FILTER (WHERE status >= 500)::numeric / NULLIF(COUNT(*), 0)) * 100,
        2
    ) AS error_rate_pct
FROM requests
WHERE timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY error_rate_pct DESC;

-- На собесе: порог CB (например 50% за 30s) ≠ error_rate в метриках (Prometheus),
-- SQL — для постмортема / дашборда, не для hot-path решения Open/Closed.
