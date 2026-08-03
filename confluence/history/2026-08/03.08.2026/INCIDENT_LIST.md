📝 ИНЦИДЕНТ СЕГОДНЯ
Что было:

Inventory пытался использовать общую БД с Shop (5465)

Не было отдельной БД для Inventory

Не было отдельного Redis для Inventory

Миграция Inventory была сломана (синтаксическая ошибка в SQL)

Что сделали:

Создали отдельную БД для Inventory (5466)

Создали отдельный Redis для Inventory (6384)

Починили миграцию Inventory

Накатили миграцию вручную через psql

Обновили docker-compose до v1.0.6

📋 ИТОГОВЫЙ СТАТУС
Сервис	gRPC	Metrics	PostgreSQL	Redis
Auth	50051 ✅	9091 ✅	5460 ✅	6379 ✅
Game	50052 ✅	9092 ✅	5461 ✅	6380 ✅
Billing	50053 ✅	9093 ✅	5462 ✅	6381 ✅
Leaderboard	50054 ✅	9094 ✅	5463 ✅	6382 ✅
Profile	50060 ✅	9099 ✅	5464 ✅	—
Shop	50055 ✅	9095 ✅	5465 ✅	6383 ✅
Inventory	50059 ✅	9096 ✅	5466 ✅	6384 ✅
Gateway-1	HTTP 8081	9095	—	—
Gateway-2	HTTP 8082	9096	—	—
Gateway-3	HTTP 8083	9097	—	—
Balancer	HTTP 8079	9098	—	—

🎯 ФИНАЛЬНЫЙ ИТОГ:
Сервис	gRPC	Metrics	HTTP (Gateway)	БД	Статус
Auth	50051	9091	/api/auth/*	PostgreSQL 5460	✅
Billing	50053	9093	/api/billing/*	PostgreSQL 5462	✅
Game	50052	9092	/api/game/*	PostgreSQL 5461	✅
Leaderboard	50054	9094	/api/leaderboard	PostgreSQL 5463	✅
Profile	50060	9099	/api/profile	PostgreSQL 5464	✅
Shop	50055	9095	/api/shop/*	PostgreSQL 5465	✅
Inventory	50059	9096	/api/inventory/*	PostgreSQL 5466	✅
Gateway	-	9095-9097	8081-8083	-	✅
Balancer	-	9098	8079	-	✅