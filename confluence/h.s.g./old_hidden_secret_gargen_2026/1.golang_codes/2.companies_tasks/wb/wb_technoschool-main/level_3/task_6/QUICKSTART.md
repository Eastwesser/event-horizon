# SalesTracker - Quick Start Guide

A modern CRUD application with financial analytics, built with Go and PostgreSQL using the WBF framework.

## Features

✅ **CRUD Operations**
- Create, read, update, delete transactions
- Support for income and expense types
- Validation of data (amount, date, category)

✅ **Analytics**
- Sum, average, count
- Median and 90th percentile calculations
- Category-based aggregation
- Date range filtering

✅ **Web Interface**
- Modern, responsive UI
- Real-time transaction management
- Interactive analytics dashboard
- CSV export functionality

✅ **Built with WBF Framework**
- PostgreSQL connection pooling via `dbpg`
- Structured logging via `zlog`
- Configuration management via `config`

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15 (or use docker-compose)

## Installation & Running

### 1. Start PostgreSQL

```bash
make docker-up
```

Or manually:
```bash
docker-compose up -d
```

### 2. Run the application

```bash
go run main.go
```

Or build and run:
```bash
make build
./task_6.exe
```

### 3. Open in browser

Navigate to: **http://localhost:8080**

## API Endpoints

### CRUD Operations

```
POST   /api/items              - Create transaction
GET    /api/items              - List transactions (with filters)
GET    /api/items/:id          - Get transaction by ID
PUT    /api/items/:id          - Update transaction
DELETE /api/items/:id          - Delete transaction
```

### Analytics

```
GET    /api/analytics          - Get overall analytics (sum, avg, median, p90)
GET    /api/analytics/categories - Get analytics by category
```

### Export

```
GET    /api/export/csv         - Export transactions to CSV
```

## Query Parameters

### For GET /api/items
- `from` - Start date (YYYY-MM-DD)
- `to` - End date (YYYY-MM-DD)
- `category` - Filter by category
- `type` - Filter by type (income/expense)
- `sort` - Sort by field (default: date)
- `order` - ASC or DESC (default: DESC)
- `limit` - Number of records
- `offset` - Pagination offset

### For /api/analytics and /api/export/csv
- `from` - Start date (YYYY-MM-DD)
- `to` - End date (YYYY-MM-DD)

## Example Requests

### Create Transaction
```bash
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{
    "type": "income",
    "amount": 5000,
    "category": "Salary",
    "date": "2024-01-15",
    "comment": "January salary"
  }'
```

### List Transactions with Filters
```bash
curl "http://localhost:8080/api/items?from=2024-01-01&to=2024-01-31&type=expense"
```

### Get Analytics
```bash
curl "http://localhost:8080/api/analytics?from=2024-01-01&to=2024-01-31"
```

### Download CSV
```bash
curl "http://localhost:8080/api/export/csv?from=2024-01-01&to=2024-01-31" > report.csv
```

## Project Structure

```
task_6/
├── config/          - Configuration management
├── handlers/        - HTTP handlers
├── models/          - Data models
├── service/         - Business logic
├── storage/         - Database operations
├── static/          - Frontend (HTML/JS)
├── main.go          - Application entry point
├── go.mod           - Go module definition
└── docker-compose.yml - PostgreSQL container config
```

## Making Changes

If you modify the code:

1. Rebuild: `make build`
2. Restart: `make clean && make run`

## Stopping

```bash
# Stop PostgreSQL
make docker-down

# Or manually
docker-compose down
```

## Environment Variables

- `DATABASE_DSN` - PostgreSQL connection string (default: postgres://postgres:8832@localhost:5442/sales_tracker?sslmode=disable)
- `SERVER_PORT` - Port to run on (default: 8080)
- `SERVER_HOST` - Host to bind to (default: localhost)

## Database Schema

The application automatically creates the following table on startup:

```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    category VARCHAR(100) NOT NULL,
    date TIMESTAMP NOT NULL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Indexes are automatically created for:
- date
- category
- type

## Performance Notes

- The application uses PERCENTILE_CONT for accurate median and percentile calculations
- Queries are optimized with indexes
- Connection pooling is configured with 10 max connections, 5 idle

## Troubleshooting

### Database connection failed
- Ensure PostgreSQL is running: `docker-compose logs postgres`
- Check DATABASE_DSN environment variable
- Verify database credentials

### Port 8080 already in use
- Change port via SERVER_PORT env var
- Or kill the process using the port

### CORS errors
- CORS is enabled for all origins in development mode
- Modify `corsMiddleware()` in main.go to restrict origins

## Features Implemented

✅ Basic CRUD (Create, Read, Update, Delete)
✅ Input validation with error handling
✅ SQL injection prevention via parameterized queries
✅ Date range filtering
✅ Category and type filtering
✅ Sorting and pagination
✅ Analytics calculations:
  - Sum
  - Average
  - Count
  - Median
  - 90th Percentile
✅ Category-based aggregation
✅ CSV export
✅ Modern web interface
✅ Responsive design
✅ Real-time updates
✅ WBF framework integration

## Next Steps (Optional Enhancements)

- [ ] Authentication and authorization
- [ ] User management (multi-user support)
- [ ] Recurring transactions
- [ ] Budget planning and alerts
- [ ] Mobile app
- [ ] Advanced reporting
- [ ] Database migrations
- [ ] API documentation (Swagger)
- [ ] Unit tests

