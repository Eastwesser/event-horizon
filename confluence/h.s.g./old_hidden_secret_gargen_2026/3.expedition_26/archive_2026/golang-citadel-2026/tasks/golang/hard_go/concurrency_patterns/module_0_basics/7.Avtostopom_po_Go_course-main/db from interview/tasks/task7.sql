CREATE TABLE customer
(
    id      INTEGER PRIMARY KEY,
    email   VARCHAR (100) NOT NULL,
    country CHAR(2) NOT NULL
);

CREATE TABLE cart_item
(
    id          INTEGER PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    title       VARCHAR (20) NOT NULL,
    amount      INTEGER NOT NULL,
    price       INTEGER NOT NULL
);

-- Вывести построчно всех кастомеров (id, email)
-- и все элементы корзины пользователя (title, amount)


-- Вывести топ-10 клиентов (id, email) по общей стоимости товаров в корзине


-- Интересуют только покупатели из России,
-- которые положили в корзину товаров не менее чем на 1000 рублей

-- Будет ли применяться индекс?
CREATE INDEX myIdx ON carts (sku, country, customer_id);
SELECT * FROM carts WHERE sku = 192 AND country = 'ru';

-- А так?
CREATE INDEX myIdx ON carts(sku, country, customer_id);
SELECT * FROM carts WHERE country = 'ru' AND customer_id = 10