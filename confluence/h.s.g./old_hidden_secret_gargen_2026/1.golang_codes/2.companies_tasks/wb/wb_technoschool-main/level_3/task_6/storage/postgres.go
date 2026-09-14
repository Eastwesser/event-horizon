package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/wb-go/wb_technoschool/level_3/task_6/models"
	"github.com/wb-go/wbf/dbpg"
)

type Storage struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		type VARCHAR(20) NOT NULL,
		amount DECIMAL(15, 2) NOT NULL,
		category VARCHAR(100) NOT NULL,
		date TIMESTAMP NOT NULL,
		comment TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
	CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);
	CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
	`

	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *Storage) CreateTransaction(ctx context.Context, tx *models.Transaction) (int64, error) {
	query := `
	INSERT INTO transactions (type, amount, category, date, comment)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`

	var id int64
	err := s.db.QueryRowContext(ctx, query, tx.Type, tx.Amount, tx.Category, tx.Date, tx.Comment).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	return id, nil
}

func (s *Storage) GetTransaction(ctx context.Context, id int64) (*models.Transaction, error) {
	query := `
	SELECT id, type, amount, category, date, comment
	FROM transactions
	WHERE id = $1
	`

	tx := &models.Transaction{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.Type,
		&tx.Amount,
		&tx.Category,
		&tx.Date,
		&tx.Comment,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

func (s *Storage) ListTransactions(ctx context.Context, filters ListFilters) ([]models.Transaction, error) {
	query := `
	SELECT id, type, amount, category, date, comment
	FROM transactions
	WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filters.FromDate != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, filters.FromDate)
		argNum++
	}

	if filters.ToDate != nil {
		query += fmt.Sprintf(" AND date < $%d", argNum)
		args = append(args, filters.ToDate)
		argNum++
	}

	if filters.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argNum)
		args = append(args, filters.Category)
		argNum++
	}

	if filters.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argNum)
		args = append(args, filters.Type)
		argNum++
	}

	if filters.SortBy == "" {
		filters.SortBy = "date"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "DESC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", pq.QuoteIdentifier(filters.SortBy), filters.SortOrder)

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filters.Limit)
		argNum++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filters.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(&tx.ID, &tx.Type, &tx.Amount, &tx.Category, &tx.Date, &tx.Comment)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

func (s *Storage) UpdateTransaction(ctx context.Context, id int64, tx *models.Transaction) error {
	query := `
	UPDATE transactions
	SET type = $1, amount = $2, category = $3, date = $4, comment = $5
	WHERE id = $6
	`

	result, err := s.db.ExecContext(ctx, query, tx.Type, tx.Amount, tx.Category, tx.Date, tx.Comment, id)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

func (s *Storage) DeleteTransaction(ctx context.Context, id int64) error {
	query := `DELETE FROM transactions WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

func (s *Storage) GetAnalytics(ctx context.Context, fromDate, toDate time.Time) (*models.Analytics, error) {
	query := `
	SELECT
		SUM(amount) as sum,
		AVG(amount) as avg,
		COUNT(*) as count,
		PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount) as median,
		PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount) as p90
	FROM transactions
	WHERE date >= $1 AND date < $2
	`

	analytics := &models.Analytics{
		MinDate: fromDate.Format("2006-01-02"),
		MaxDate: toDate.Format("2006-01-02"),
	}

	var sum, avg, median, p90 sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, fromDate, toDate).Scan(
		&sum,
		&avg,
		&analytics.Count,
		&median,
		&p90,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	if sum.Valid {
		analytics.Sum = sum.Float64
	}
	if avg.Valid {
		analytics.Avg = avg.Float64
	}
	if median.Valid {
		analytics.Median = median.Float64
	}
	if p90.Valid {
		analytics.P90 = p90.Float64
	}

	return analytics, nil
}

func (s *Storage) GetCategoryAnalytics(ctx context.Context, fromDate, toDate time.Time) ([]models.CategoryAnalytics, error) {
	query := `
	SELECT
		category,
		SUM(amount) as sum,
		COUNT(*) as count,
		AVG(amount) as avg
	FROM transactions
	WHERE date >= $1 AND date < $2
	GROUP BY category
	ORDER BY sum DESC
	`

	rows, err := s.db.QueryContext(ctx, query, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get category analytics: %w", err)
	}
	defer rows.Close()

	var analytics []models.CategoryAnalytics
	for rows.Next() {
		var ca models.CategoryAnalytics
		err := rows.Scan(&ca.Category, &ca.Sum, &ca.Count, &ca.Avg)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category analytics: %w", err)
		}
		analytics = append(analytics, ca)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category analytics: %w", err)
	}

	return analytics, nil
}

func (s *Storage) ExportToCSV(ctx context.Context, fromDate, toDate time.Time) (string, error) {
	filters := ListFilters{
		FromDate: &fromDate,
		ToDate:   &toDate,
		Limit:    10000,
	}

	transactions, err := s.ListTransactions(ctx, filters)
	if err != nil {
		return "", err
	}

	// Build CSV
	var sb strings.Builder
	sb.WriteString("ID,Type,Amount,Category,Date,Comment\n")

	for _, tx := range transactions {
		// Escape comment for CSV
		comment := strings.ReplaceAll(tx.Comment, "\"", "\"\"")
		if strings.ContainsAny(comment, ",\"\n") {
			comment = "\"" + comment + "\""
		}

		line := fmt.Sprintf("%d,%s,%.2f,%s,%s,%s\n",
			tx.ID,
			tx.Type,
			tx.Amount,
			tx.Category,
			tx.Date.Format("2006-01-02"),
			comment,
		)
		sb.WriteString(line)
	}

	return sb.String(), nil
}

type ListFilters struct {
	FromDate  *time.Time
	ToDate    *time.Time
	Category  string
	Type      string
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

