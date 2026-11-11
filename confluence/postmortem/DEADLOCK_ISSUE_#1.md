DB deadlock problem (commented out query is the deadlock)

err = tx.QueryRowContext(ctx, `
		SELECT id, price, refundable_until, fulfilled_at
		FROM purchases
		WHERE user_id = $1 AND item_id = $2
		  AND status = 'COMPLETED' AND refunded_at IS NULL
		ORDER BY COALESCE(completed_at, purchased_at) DESC
		LIMIT 1
	`, userID, itemID).Scan(&purchaseID, &price, &refundableUntil, &fulfilledAt)
	// SELECT p.id, p.price, p.refundable_until, p.fulfilled_at, i.category
	// 	FROM purchases p
	// 	JOIN items i ON i.id = p.item_id
	// 	WHERE p.user_id = $1 AND p.item_id = $2
	// 	  AND p.status = 'COMPLETED' AND p.refunded_at IS NULL
	// 	ORDER BY COALESCE(p.completed_at, p.purchased_at) DESC
	// 	LIMIT 1
	// 	FOR UPDATE OF p

Почему я убрал JOIN items — отвечаю на твой вопрос
Ты спросил: «почему убрал p.id, написал просто id? Это чтобы не тормозило?»

Две причины, обе важные:

1. JOIN items + FOR UPDATE OF p — скрытый лок на items
Postgres, когда видит JOIN items i ON i.id = p.item_id с FOR UPDATE OF p, всё равно читает items (для JOIN). FOR UPDATE OF p говорит «блокируй только p», но план запроса может быть таким:

Nested Loop с purchases как outer, items как inner → сначала лок на purchases, потом read items.

Nested Loop с items как outer, purchases как inner → сначала read items, потом лок на purchases.

Hash Join → сначала сканирует обе таблицы, потом лок на purchases.

Порядок локов зависит от плана, а план зависит от статистики, размера таблиц, work_mem, версии Postgres. То есть порядок локов непредсказуем. Это хуже, чем фиксированный неправильный порядок — потому что воспроизвести deadlock сложно, а случиться может на любом плане.

Убрав JOIN, я убрал неявный лок на items из этого запроса. Теперь items лочится только явным SELECT 1 FROM items WHERE id = $1 FOR UPDATE (шаг 1) — всегда первым, всегда предсказуемо.

2. FOR UPDATE OF p в JOIN — не тот лок
FOR UPDATE OF p в запросе с JOIN работает только для p, но не для i. Если items нужен лок (а он нужен — мы потом делаем UPDATE items SET stock = stock + 1 в шаге 6), то FOR UPDATE OF p не даёт его. Мы бы взяли items лок позже (в шаге 6, UPDATE items), а это снова не тот порядок.

Убрав JOIN и добавив явный SELECT 1 FROM items ... FOR UPDATE (шаг 1), я гарантирую: items лочится первым, всегда.

3. «Чтобы не тормозило» — тоже да, но не главное
JOIN на purchases × items под 20 VU — это лишний read items (индексный, но всё равно). Отдельный SELECT category FROM items WHERE id = $1 (шаг 3) дешевле, потому что items уже залочен (шаг 1), и это тот же план, что и в purchase. Но главное — предсказуемость локов, а не скорость.

4. Почему id, а не p.id
Потому что p больше нет — нет алиаса. FROM purchases без алиаса → id, price, ... однозначны. Это не про скорость, это про чистоту: если нет JOIN, нет смысла в алиасе p.