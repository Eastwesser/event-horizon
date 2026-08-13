package repository

import (
    "context"
	"errors"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
	
	"github.com/jackc/pgx/v5"
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

// AddBalance credits currency atomically via an upsert-increment (INSERT ... ON CONFLICT
// DO UPDATE SET balance = balance + amount), which avoids the classic
// SELECT-then-UPDATE lost-update race that the old implementation had: two concurrent
// credits could both read the same stale balance and one would silently overwrite the
// other. The balance change, transaction row, and outbox event are all written in one
// transaction so an event is never lost or emitted without the balance actually changing.
func (r *PostgresBillingRepo) AddBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error) {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback(ctx)

    var newBalance int
    upsertQuery := `
        INSERT INTO user_currencies (user_id, currency_type, balance, version)
        VALUES ($1, $2, $3, 1)
        ON CONFLICT (user_id, currency_type)
        DO UPDATE SET
            balance = user_currencies.balance + EXCLUDED.balance,
            version = user_currencies.version + 1,
            updated_at = NOW()
        RETURNING balance
    `
    if err := tx.QueryRow(ctx, upsertQuery, userID, currency, amount).Scan(&newBalance); err != nil {
        return 0, err
    }

    txQuery := `INSERT INTO transactions (user_id, currency_type, amount, balance_after, reason, reference_id) 
                VALUES ($1, $2, $3, $4, $5, $6)`
    if _, err = tx.Exec(ctx, txQuery, userID, currency, amount, newBalance, reason, referenceID); err != nil {
        return 0, err
    }

    if err := insertOutboxEvent(ctx, tx, "balance.updated", balanceUpdatedEvent{
        UserID: userID, Currency: string(currency), Delta: amount, Balance: newBalance, Reason: reason,
    }); err != nil {
        return 0, err
    }

    if err := tx.Commit(ctx); err != nil {
        return 0, err
    }

    return newBalance, nil
}

// SpendBalance debits currency atomically via a conditional UPDATE
// (`SET balance = balance - $1 WHERE balance >= $1`). Postgres re-checks the WHERE
// clause against the row's current committed value when acquiring the row lock for the
// UPDATE, so this can never drive the balance negative or lose a concurrent update —
// unlike the previous SELECT-then-UPDATE approach.
func (r *PostgresBillingRepo) SpendBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error) {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback(ctx)

    var newBalance int
    updateQuery := `
        UPDATE user_currencies
        SET balance = balance - $1, version = version + 1, updated_at = NOW()
        WHERE user_id = $2 AND currency_type = $3 AND balance >= $1
        RETURNING balance
    `
    err = tx.QueryRow(ctx, updateQuery, amount, userID, currency).Scan(&newBalance)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            var currentBalance int
            checkErr := tx.QueryRow(ctx,
                `SELECT balance FROM user_currencies WHERE user_id = $1 AND currency_type = $2`,
                userID, currency).Scan(&currentBalance)
            if checkErr != nil {
                return 0, fmt.Errorf("user not found or no balance: %w", checkErr)
            }
            return 0, fmt.Errorf("insufficient balance: have %d, need %d", currentBalance, amount)
        }
        return 0, err
    }

    txQuery := `INSERT INTO transactions (user_id, currency_type, amount, balance_after, reason, reference_id) 
                VALUES ($1, $2, $3, $4, $5, $6)`
    if _, err = tx.Exec(ctx, txQuery, userID, currency, -amount, newBalance, reason, referenceID); err != nil {
        return 0, err
    }

    if err := insertOutboxEvent(ctx, tx, "balance.updated", balanceUpdatedEvent{
        UserID: userID, Currency: string(currency), Delta: -amount, Balance: newBalance, Reason: reason,
    }); err != nil {
        return 0, err
    }

    if err := tx.Commit(ctx); err != nil {
        return 0, err
    }

    return newBalance, nil
}

type balanceUpdatedEvent struct {
    UserID   string `json:"user_id"`
    Currency string `json:"currency"`
    Delta    int    `json:"delta"`
    Balance  int    `json:"balance"`
    Reason   string `json:"reason"`
}

func insertOutboxEvent(ctx context.Context, tx pgx.Tx, eventType string, event any) error {
    payload, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal outbox payload: %w", err)
    }
    _, err = tx.Exec(ctx, `INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`, eventType, payload)
    return err
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

// BatchInsertTransactions массовая вставка транзакций
func (r *PostgresBillingRepo) BatchInsertTransactions(ctx context.Context, transactions []Transaction) error {
    if len(transactions) == 0 {
        return nil
    }
    
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    // Используем COPY для массовой вставки
    copyFrom := pgx.CopyFromSlice(len(transactions), func(i int) ([]any, error) {
        t := transactions[i]
        return []any{
            t.ID, t.UserID, t.Currency, t.Amount,
            t.BalanceAfter, t.Reason, t.ReferenceID, t.CreatedAt,
        }, nil
    })
    
    _, err = tx.CopyFrom(ctx, pgx.Identifier{"transactions"},
        []string{"id", "user_id", "currency_type", "amount", "balance_after", "reason", "reference_id", "created_at"},
        copyFrom)
    if err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
