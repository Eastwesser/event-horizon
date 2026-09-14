-- Вывести всех кастомеров и их товары в корзине

SELECT c.id, c.email, ci.title, ci.amount
FROM customer c
LEFT JOIN cart_item ci ON c.id = ci.customer_id
ORDER BY c.id;

-- LEFT JOIN чтобы вывести всех пользователей, даже если у кого-то пустая корзина.

---

-- Топ-10 клиентов по общей стоимости товаров в корзине

SELECT c.id, c.email, SUM(ci.amount * ci.price) AS total_spent
FROM customer c
JOIN cart_item ci ON c.id = ci.customer_id
GROUP BY c.id, c.email
ORDER BY total_spent DESC
LIMIT 10;

---

-- Покупатели из России с корзиной от 1000 рублей

SELECT c.id, c.email, SUM(ci.amount * ci.price) AS total_sum
FROM customer c
JOIN cart_item ci ON c.id = ci.customer_id
WHERE c.country = 'RU'
GROUP BY c.id, c.email
HAVING SUM(ci.amount * ci.price) >= 1000;


---

-- Будет ли применяться индекс?

CREATE INDEX myIdx ON carts (sku, country, customer_id);
SELECT * FROM carts WHERE sku = 192 AND country = 'ru';

-- индекс по (sku, country, customer_id) будет использоваться для фильтрации по (sku, country), даже если customer_id не указан.

---

CREATE INDEX myIdx ON carts (sku, country, customer_id);
SELECT * FROM carts WHERE country = 'ru' AND customer_id = 10;

-- myIdx(sku, country, customer_id) не будет использоваться, потому что:
-- Первый столбец в индексе — `sku`, но в запросе его нет.