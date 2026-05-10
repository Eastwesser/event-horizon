package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type CurrencyType string

const (
    Lamps   CurrencyType = "lamps"
    Tickets CurrencyType = "tickets"
)

type Balance struct {
    UserID     string
    Currency   CurrencyType
    Balance    int
    UpdatedAt  time.Time
}

type Transaction struct {
    ID           string
    UserID       string
    Currency     CurrencyType
    Amount       int
    BalanceAfter int
    Reason       string
    ReferenceID  string
    CreatedAt    time.Time
}

type BillingRepository interface {
    GetBalance(ctx context.Context, userID string, currency CurrencyType) (int, error)
    AddBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error)
    SpendBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error)
    GetTransactionHistory(ctx context.Context, userID string, currency CurrencyType, limit, offset int) ([]Transaction, int, error)
}

type PostgresBillingRepo struct {
    db *pgxpool.Pool
}

func NewPostgresBillingRepo(db *pgxpool.Pool) *PostgresBillingRepo {
    return &PostgresBillingRepo{db: db}
}

func (r *PostgresBillingRepo) GetBalance(ctx context.Context, userID string, currency CurrencyType) (int, error) {
    var balance int
    query := `SELECT balance FROM user_currencies WHERE user_id = $1 AND currency_type = $2`
    err := r.db.QueryRow(ctx, query, userID, currency).Scan(&balance)
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, nil // новый пользователь, баланс 0
        }
        return 0, err
    }
    return balance, nil
}

func (r *PostgresBillingRepo) AddBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error) {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback(ctx)

    // Получаем текущий баланс
    var currentBalance int
    query := `SELECT balance FROM user_currencies WHERE user_id = $1 AND currency_type = $2`
    err = tx.QueryRow(ctx, query, userID, currency).Scan(&currentBalance)
    if err != nil && err != sql.ErrNoRows {
        return 0, err
    }

    newBalance := currentBalance + amount

    if currentBalance == 0 && err == sql.ErrNoRows {
        // Вставляем новую запись
        insertQuery := `INSERT INTO user_currencies (user_id, currency_type, balance) VALUES ($1, $2, $3)`
        _, err = tx.Exec(ctx, insertQuery, userID, currency, newBalance)
    } else {
        // Обновляем существующую
        updateQuery := `UPDATE user_currencies SET balance = $1, updated_at = NOW() WHERE user_id = $2 AND currency_type = $3`
        _, err = tx.Exec(ctx, updateQuery, newBalance, userID, currency)
    }
    if err != nil {
        return 0, err
    }

    // Записываем транзакцию
    txQuery := `INSERT INTO transactions (user_id, currency_type, amount, balance_after, reason, reference_id) 
                VALUES ($1, $2, $3, $4, $5, $6)`
    _, err = tx.Exec(ctx, txQuery, userID, currency, amount, newBalance, reason, referenceID)
    if err != nil {
        return 0, err
    }

    if err := tx.Commit(ctx); err != nil {
        return 0, err
    }

    return newBalance, nil
}

func (r *PostgresBillingRepo) SpendBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error) {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback(ctx)

    // Проверяем баланс
    var currentBalance int
    query := `SELECT balance FROM user_currencies WHERE user_id = $1 AND currency_type = $2`
    err = tx.QueryRow(ctx, query, userID, currency).Scan(&currentBalance)
    if err != nil {
        return 0, fmt.Errorf("user not found or no balance: %w", err)
    }

    if currentBalance < amount {
        return 0, fmt.Errorf("insufficient balance: have %d, need %d", currentBalance, amount)
    }

    newBalance := currentBalance - amount

    updateQuery := `UPDATE user_currencies SET balance = $1, updated_at = NOW() WHERE user_id = $2 AND currency_type = $3`
    _, err = tx.Exec(ctx, updateQuery, newBalance, userID, currency)
    if err != nil {
        return 0, err
    }

    txQuery := `INSERT INTO transactions (user_id, currency_type, amount, balance_after, reason, reference_id) 
                VALUES ($1, $2, $3, $4, $5, $6)`
    _, err = tx.Exec(ctx, txQuery, userID, currency, -amount, newBalance, reason, referenceID)
    if err != nil {
        return 0, err
    }

    if err := tx.Commit(ctx); err != nil {
        return 0, err
    }

    return newBalance, nil
}

func (r *PostgresBillingRepo) GetTransactionHistory(ctx context.Context, userID string, currency CurrencyType, limit, offset int) ([]Transaction, int, error) {
    var total int
    countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND currency_type = $2`
    err := r.db.QueryRow(ctx, countQuery, userID, currency).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    rows, err := r.db.Query(ctx, `
        SELECT id, user_id, currency_type, amount, balance_after, reason, reference_id, created_at
        FROM transactions 
        WHERE user_id = $1 AND currency_type = $2
        ORDER BY created_at DESC
        LIMIT $3 OFFSET $4
    `, userID, currency, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var transactions []Transaction
    for rows.Next() {
        var t Transaction
        err := rows.Scan(&t.ID, &t.UserID, &t.Currency, &t.Amount, &t.BalanceAfter, &t.Reason, &t.ReferenceID, &t.CreatedAt)
        if err != nil {
            return nil, 0, err
        }
        transactions = append(transactions, t)
    }

    return transactions, total, nil
}
