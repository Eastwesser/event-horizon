-- Запрос 1:

SELECT u.user_id, u.name, p.purchase_id
FROM users u
JOIN purchases p ON u.user_id = p.user_id
LEFT JOIN banlist b ON u.user_id = b.user_id
WHERE b.ban_date IS NULL OR p.purchase_date < b.ban_date;

-- Запрос 2:

SELECT u.user_id, u.name, SUM(p.amount) AS total_spent
FROM users u
JOIN purchases p ON u.user_id = p.user_id
GROUP BY u.user_id, u.name
HAVING SUM(p.amount) > 5000;