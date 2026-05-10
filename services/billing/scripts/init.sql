-- Таблица балансов пользователей
CREATE TABLE IF NOT EXISTS user_currencies (
    user_id UUID NOT NULL,
    currency_type VARCHAR(20) NOT NULL, -- 'lamps' или 'tickets'
    balance INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, currency_type)
);

-- Таблица транзакций (для истории и отладки)
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    currency_type VARCHAR(20) NOT NULL,
    amount INT NOT NULL,
    balance_after INT NOT NULL,
    reason VARCHAR(50) NOT NULL,
    reference_id VARCHAR(100) UNIQUE, -- idempotency key
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_reference_id ON transactions(reference_id);

-- Инициализация баланса для нового пользователя (триггер или отдельная логика)