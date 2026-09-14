# SQL_JOIN

Скрининг SQL: JOIN / LEFT JOIN, фильтр по дате до бана, DISTINCT user+SKU, SUM с GROUP BY.

Классика из урока 00: покупки до попадания в `ban_list`, пользователи с суммой > 5000.

Что сказать: «INNER когда обе стороны обязательны; LEFT + `IS NULL OR date < ban` для «не забанен или купил раньше»».

Разбор/SQL: [`../../../../../00_4th_March_2026/`](../../../../../00_4th_March_2026/) (`05_sql_purchases_before_ban`).  
Отдельного Go pack в `ozon/code` нет.
