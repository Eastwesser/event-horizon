# SalesTracker

**A comprehensive financial tracking and analytics service built with Go**

## Overview

SalesTracker is a modern CRUD application designed to manage financial transactions (sales, expenses, income) with built-in analytics capabilities. It demonstrates professional Go development practices including proper project structure, database optimization, and user-friendly web interface.

## 🎯 Problem Statement

Many services need to:
- Store and manage transaction records
- Perform quick CRUD operations
- Analyze data trends (sum, average, median, percentiles)
- Filter and sort results efficiently
- Export data for external use

SalesTracker solves all these problems in an elegant, performant way.

## ✨ Key Features

### CRUD Operations
- **POST /api/items** - Create new transaction
- **GET /api/items** - List transactions with advanced filtering
- **GET /api/items/{id}** - Retrieve specific transaction
- **PUT /api/items/{id}** - Update transaction
- **DELETE /api/items/{id}** - Delete transaction

### Advanced Analytics
- **Sum** - Total amount for the period
- **Average** - Mean transaction value
- **Count** - Number of transactions
- **Median** - Middle value (SQL PERCENTILE_CONT)
- **90th Percentile** - Upper quartile analysis
- **Category Analytics** - Breakdown by category

### Data Management
- **Filtering** - By date range, category, type
- **Sorting** - Customizable order and field
- **Pagination** - Limit and offset support
- **CSV Export** - Download reports for external analysis

### Modern Web Interface
- Responsive design works on all devices
- Real-time transaction management
- Interactive analytics dashboard
- One-click CSV export
- Beautiful UI with gradient theme

## 🛠️ Technology Stack

| Component | Technology |
|-----------|-----------|
| **Language** | Go 1.21+ |
| **Web Framework** | Gin |
| **Database** | PostgreSQL 15 |
| **Framework** | WBF (internal) |
| **Connection Pooling** | dbpg (WBF module) |
| **Logging** | zlog (WBF module) |
| **Configuration** | config (WBF module) |
| **Frontend** | HTML5 + Vanilla JavaScript |
| **Containerization** | Docker & Docker Compose |

## 📦 Project Structure

```
task_6/
├── config/
│   └── config.go              # Configuration management
├── handlers/
│   └── handlers.go            # HTTP request handlers
├── models/
│   └── transaction.go         # Data models and validation
├── service/
│   └── service.go             # Business logic layer
├── storage/
│   └── postgres.go            # Database operations
├── static/
│   └── index.html             # Web interface
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── docker-compose.yml         # PostgreSQL container
├── Makefile                   # Build commands
├── env.example                # Environment variables template
├── QUICKSTART.md              # Quick start guide
└── README.md                  # This file
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker & Docker Compose
- PostgreSQL 15 (or use Docker)

### Installation

1. **Start PostgreSQL**
```bash
make docker-up
# or
docker-compose up -d
```

This will start PostgreSQL on port **5442** with password **8832**.

2. **Run Application**
```bash
go run main.go
# or use pre-built binary
./task_6.exe
```

3. **Open in Browser**
Navigate to: http://localhost:8080

## 📊 API Usage Examples

### Create Transaction
```bash
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{
    "type": "income",
    "amount": 5000.00,
    "category": "Salary",
    "date": "2024-01-15",
    "comment": "January salary"
  }'
```

### List Transactions
```bash
# All transactions
curl http://localhost:8080/api/items

# Filtered by date range
curl "http://localhost:8080/api/items?from=2024-01-01&to=2024-01-31"

# By category
curl "http://localhost:8080/api/items?category=Food"

# By type
curl "http://localhost:8080/api/items?type=expense"
```

### Get Analytics
```bash
curl "http://localhost:8080/api/analytics?from=2024-01-01&to=2024-01-31"
```

Response:
```json
{
  "sum": 15000.50,
  "avg": 1500.05,
  "count": 10,
  "median": 1200.00,
  "p90": 3500.00,
  "min_date": "2024-01-01",
  "max_date": "2024-01-31"
}
```

### Category Analytics
```bash
curl "http://localhost:8080/api/analytics/categories?from=2024-01-01&to=2024-01-31"
```

### Export to CSV
```bash
curl "http://localhost:8080/api/export/csv?from=2024-01-01&to=2024-01-31" \
  -o report.csv
```

## 🔍 Database Schema

```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,           -- 'income' or 'expense'
    amount DECIMAL(15, 2) NOT NULL,      -- Transaction amount
    category VARCHAR(100) NOT NULL,      -- Classification
    date TIMESTAMP NOT NULL,              -- Transaction date
    comment TEXT,                         -- Optional notes
    created_at TIMESTAMP DEFAULT NOW()
);

-- Optimized indexes for common queries
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_category ON transactions(category);
CREATE INDEX idx_transactions_type ON transactions(type);
```

## 🎨 Frontend Features

- **Add Transaction Form** - Easy data entry with validation
- **Transaction Table** - Sortable, filterable list
- **Edit/Delete Actions** - Inline transaction management
- **Advanced Filters** - By date range, category, type
- **Analytics Dashboard** - Visual statistics
- **Category Breakdown** - Detailed spending by category
- **CSV Export** - One-click data export
- **Responsive Design** - Mobile-friendly interface

## 🔐 Security Features

✅ **SQL Injection Prevention**
- Parameterized queries using lib/pq
- All user input properly escaped

✅ **Input Validation**
- Amount must be non-negative
- Date validation
- Category non-empty requirement
- Type must be 'income' or 'expense'

✅ **CORS Enabled**
- Cross-origin requests allowed
- Configurable via middleware

## 📈 Performance Optimizations

- **Connection Pooling** - 10 max open, 5 idle connections
- **Database Indexes** - On date, category, type columns
- **Query Optimization** - Efficient SQL with PERCENTILE_CONT
- **Pagination Support** - Limit/offset for large datasets

## 🛠️ Development Commands

```bash
# Start PostgreSQL
make docker-up

# Stop PostgreSQL
make docker-down

# View logs
make docker-logs

# Run application
make run

# Build executable
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## 📝 Data Models

### Transaction
```go
type Transaction struct {
    ID       int64           // Unique identifier
    Type     TransactionType // "income" or "expense"
    Amount   float64         // Transaction amount
    Category string          // Classification
    Date     time.Time       // Transaction date
    Comment  string          // Optional notes
}
```

### Analytics
```go
type Analytics struct {
    Sum      float64 // Total amount
    Avg      float64 // Average amount
    Count    int64   // Number of transactions
    Median   float64 // Median value
    P90      float64 // 90th percentile
    MinDate  string  // Period start
    MaxDate  string  // Period end
}
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific test
go test ./service/...
```

## 📚 WBF Framework Integration

This project uses the internal WBF framework for:

- **dbpg** - PostgreSQL connection management with retry logic
- **zlog** - Structured logging via zerolog
- **config** - Configuration management with Viper
- **ginext** - Gin web framework extensions

See the [WBF documentation](https://github.com/wb-go/wbf) for more details.

## 🔄 Workflow

1. **Add Transactions** - Enter income/expense records
2. **View & Filter** - Browse transactions with custom filters
3. **Analyze** - View statistics and trends
4. **Export** - Download CSV reports
5. **Update/Delete** - Modify or remove transactions as needed

## 🐛 Troubleshooting

### Database Connection Error
```
Failed to connect to database
```
**Solution:** Ensure PostgreSQL is running
```bash
docker-compose logs postgres
```

**Database Connection Details:**
- Host: localhost
- Port: 5442
- User: postgres
- Password: 8832
- Database: sales_tracker

### Port Already in Use
```
listen tcp: address already in use
```
**Solution:** Change port via environment variable
```bash
SERVER_PORT=9000 go run main.go
```

### CORS Errors in Browser
**Solution:** CORS is already enabled. If you see CORS errors, check that the API is responding correctly.

## 📋 Implementation Checklist

- ✅ CRUD operations (Create, Read, Update, Delete)
- ✅ HTTP API endpoints
- ✅ PostgreSQL integration
- ✅ Input validation
- ✅ SQL injection prevention
- ✅ Analytics (sum, avg, count, median, p90)
- ✅ Date range filtering
- ✅ Category aggregation
- ✅ CSV export
- ✅ Web interface
- ✅ Responsive design
- ✅ Modern UI with gradients
- ✅ Real-time updates
- ✅ WBF framework usage

## 🎓 Learning Outcomes

This project demonstrates:
- Professional Go project structure
- REST API design best practices
- PostgreSQL query optimization
- HTML5 & Vanilla JavaScript
- Error handling and validation
- Configuration management
- Docker containerization
- Database connection pooling
- Structured logging
- Clean code architecture

## 📄 License

Project completed as part of WB TechnoSchool level 3 tasks.

## 🤝 Contributing

This is a learning project. Feel free to fork, modify, and improve!

## 📞 Support

For issues or questions:
1. Check QUICKSTART.md for common issues
2. Review error logs: `docker-compose logs postgres`
3. Verify database connection string
4. Check port availability

---

**Built with ❤️ using Go and WBF Framework**
