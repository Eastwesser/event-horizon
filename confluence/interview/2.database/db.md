🗄️ 2. БАЗЫ ДАННЫХ


❓ Какой индекс для поиска по user_id и created_at?
    Ответ: Создам составной индекс:

    sql
    CREATE INDEX idx_user_created ON orders(user_id, created_at DESC);
    Почему:

    Запросы обычно фильтруют по user_id и сортируют по created_at

    Индекс покрывает оба условия

    DESC для сортировки по убыванию (самые новые сначала)

    В Event Horizon: У нас есть аналогичный индекс в Billing:

    sql
    CREATE INDEX idx_user_currencies_user ON user_currencies(user_id);


❓ Как работает MVCC в PostgreSQL?
    Ответ: MVCC (Multi-Version Concurrency Control) — это механизм, который создает новую версию строки при каждом изменении, а не блокирует её.

    Как это работает:

    При UPDATE создается новая версия строки, старая остается.

    Каждая транзакция видит только те версии, которые были созданы до её начала.

    Читающие транзакции не блокируют пишущие, и наоборот.

    Влияние на производительность:

    Плюс: Высокая конкурентность, нет блокировок.

    Минус: Старые версии накапливаются, нужен VACUUM.


❓ VACUUM — нормально, VACUUM FULL — зло?
    Ответ:

    VACUUM — помечает мертвые строки как свободные, не блокирует таблицу. Нужно запускать регулярно.

    VACUUM FULL — перестраивает таблицу, освобождает место на диске, но блокирует таблицу на всё время.

    В продакшене: Использую VACUUM с autovacuum, VACUUM FULL — только в окно обслуживания.

    В Event Horizon: У нас включен autovacuum для всех таблиц.


❓ Что если закончатся XID?
    Ответ: XID (Transaction ID) — 32-битное число, максимум ~4 млрд. Если они закончатся, PostgreSQL перестанет принимать новые транзакции.

    Решение: Включить autovacuum, который периодически "замораживает" старые транзакции, сбрасывая их XID.


❓ Как хранишь геоданные? Что такое GiST?
    Ответ: В Roolz я использовал PostGIS и GiST-индекс (Generalized Search Tree) для геоданных.

    GiST — это индекс, который работает не с числами, а с геометрией. Он позволяет быстро искать точки внутри полигона или радиус.

    sql
    CREATE INDEX idx_location ON points USING GIST (location);


❓ Согласованность между MongoDB и PostgreSQL?
    Ответ: В Event Horizon я использую только PostgreSQL, поэтому этой проблемы нет.

    Если бы использовал две БД:

    Saga — распределенная транзакция с компенсацией.

    Eventual consistency — допускаю, что данные будут согласованы через несколько секунд.

    Если рассинхронизируются: Запускаю скрипт, который сравнивает данные и восстанавливает.


❓ Как инвалидируешь кэш в Redis?
    Ответ: В Event Horizon я использую Cache-Aside (Lazy Loading):

    При запросе сначала проверяю Redis

    Если есть — отдаю

    Если нет — иду в БД, сохраняю в Redis

    При обновлении — удаляю ключ из Redis

    Код из Billing:

    go
    func (s *billingService) AddCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error) {
        // Обновляем баланс в PostgreSQL
        newBalance, err := s.pgRepo.AddBalance(ctx, userID, currency, amount, reason, referenceID)
        if err != nil {
            return 0, err
        }
        // Инвалидируем кеш
        s.redisRepo.DeleteBalance(ctx, userID, currency)
        return newBalance, nil
    }


❓ Read-Through vs Write-Through?
    Ответ:

    Read-Through — при чтении кэш сам загружает данные из БД, если их нет.

    Write-Through — при записи данные сохраняются и в БД, и в кэш.

    В Event Horizon я использую Cache-Aside (не Read-Through), потому что:

    Проще контролировать

    Меньше нагрузка на кэш при редких запросах


❓ Как искать пользователя по email в таблице на 1 млрд строк?
    Ответ:

    Создать индекс на email:

    sql
    CREATE INDEX idx_users_email ON users(email);
    Партиционирование — разбить таблицу по email или user_id.

    В крайнем случае — использовать Elasticsearch.


❓ Что такое шардирование и когда оно нужно?
    Ответ: Шардирование — это горизонтальное разделение данных по разным серверам.

    Когда нужно:

    Таблица > 1 ТБ

    Индексы перестают помещаться в память

    RPS > 10 000 на одну БД

    Как шардировать заказы:

    sql
    -- По user_id (модуль)
    SELECT * FROM orders WHERE user_id % 4 = 0; -- в шард 0