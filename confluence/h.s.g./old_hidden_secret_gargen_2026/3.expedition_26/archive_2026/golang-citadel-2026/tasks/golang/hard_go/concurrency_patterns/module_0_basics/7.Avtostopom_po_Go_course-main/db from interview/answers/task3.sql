-- Задание 1: Вывести все айтемы с максимальной стоимостью

SELECT *
FROM example1
WHERE cost = (
    SELECT MAX(cost)
    FROM example1
);

-- Задание 2: Вывести последние комментарии каждого автора (по дате)

SELECT e2.*
FROM example2 e2
JOIN (
    SELECT author, MAX(date) AS max_date
    FROM example2
    GROUP BY author
) AS sub ON e2.author = sub.author
        AND e2.date   = sub.max_date;