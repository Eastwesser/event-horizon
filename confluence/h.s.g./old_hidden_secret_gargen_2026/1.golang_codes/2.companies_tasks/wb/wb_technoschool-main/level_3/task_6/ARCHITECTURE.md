# SalesTracker - Architecture & Design

## 📐 System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Web Browser                          │
│              (HTML + Vanilla JavaScript)                │
└────────────────────────┬────────────────────────────────┘
                         │ HTTP/REST
                         ↓
┌─────────────────────────────────────────────────────────┐
│                  Gin Web Server                         │
│              (Task 6 - main.go)                         │
│  Port: 8080 | CORS Enabled | Static Files Served       │
└────────────────────────┬────────────────────────────────┘
                         │
    ┌────────────────────┼────────────────────┐
    ↓                    ↓                    ↓
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  Handlers   │  │  Handlers   │  │  Handlers   │
│   (CRUD)    │  │ (Analytics) │  │  (Export)   │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
                        ↓
    ┌───────────────────────────────────┐
    │    Service Layer (Business Logic) │
    │   - Validation                    │
    │   - Transaction management        │
    │   - Analytics calculation         │
    └───────────────────────────────────┘
                        │
                        ↓
    ┌───────────────────────────────────┐
    │    Storage Layer (Data Access)    │
    │   - SQL Query Building            │
    │   - Database Operations           │
    │   - Transaction mapping           │
    └───────────────────────────────────┘
                        │
                        ↓
    ┌───────────────────────────────────┐
    │  PostgreSQL Database              │
    │  (Port: 5442 | Password: 8832)    │
    │  Schema: sales_tracker            │
    └───────────────────────────────────┘
```

## 📁 Project Structure

### Core Directories

```
task_6/
├── config/              # Configuration management
│   └── config.go        # Load config from env/file
│
├── handlers/            # HTTP Request Handlers
│   └── handlers.go      # CRUD & Analytics endpoints
│
├── models/              # Data Models
│   └── transaction.go   # Transaction & Analytics models + validation
│
├── service/             # Business Logic Layer
│   └── service.go       # Service methods (CRUD & analytics)
│
├── storage/             # Database Access Layer
│   └── postgres.go      # PostgreSQL operations & queries
│
├── static/              # Frontend
│   └── index.html       # Web UI (HTML + JavaScript)
│
├── main.go              # Application entry point
├── docker-compose.yml   # PostgreSQL container config
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── Makefile             # Build commands
├── env.example          # Environment variables template
│
├── QUICKSTART.md        # Quick start guide
├── INSTALLATION.md      # Installation instructions
├── ARCHITECTURE.md      # This file
└── test-api.ps1         # PowerShell API test script
```

## 🏗️ Layered Architecture

### Layer 1: Presentation Layer (Handlers)
**File:** `handlers/handlers.go`

Responsibility: HTTP request/response handling

**Endpoints:**
- `POST /api/items` - Create transaction
- `GET /api/items` - List transactions
- `GET /api/items/:id` - Get single transaction
- `PUT /api/items/:id` - Update transaction
- `DELETE /api/items/:id` - Delete transaction
- `GET /api/analytics` - Get overall analytics
- `GET /api/analytics/categories` - Category breakdown
- `GET /api/export/csv` - Export to CSV

### Layer 2: Business Logic Layer (Service)
**File:** `service/service.go`

Responsibility: Transaction validation, business rules

**Functions:**
- Validate transactions before database operation
- Call storage layer for database operations
- Apply business logic rules

### Layer 3: Data Access Layer (Storage)
**File:** `storage/postgres.go`

Responsibility: SQL queries and database operations

**Operations:**
- CRUD transactions in PostgreSQL
- Complex analytical queries
- CSV data export

### Layer 4: Database Layer
**Type:** PostgreSQL 15 Alpine

Connection details:
- Host: localhost
- Port: 5442
- User: postgres
- Password: 8832
- Database: sales_tracker

## 🗂️ Data Models

### Transaction Model

```go
type Transaction struct {
    ID       int64           // Auto-increment primary key
    Type     TransactionType // "income" or "expense"
    Amount   float64         // Transaction amount
    Category string          // Classification (e.g., Salary, Food)
    Date     time.Time       // Transaction date
    Comment  string          // Optional notes
}
```

**Validation Rules:**
- Amount must be >= 0
- Category cannot be empty
- Type must be "income" or "expense"
- Date must be valid and not in the future

### Analytics Model

```go
type Analytics struct {
    Sum      float64 // Total amount sum
    Avg      float64 // Average amount
    Count    int64   // Number of transactions
    Median   float64 // 50th percentile
    P90      float64 // 90th percentile
    MinDate  string  // Period start
    MaxDate  string  // Period end
}
```

### Category Analytics Model

```go
type CategoryAnalytics struct {
    Category string  // Category name
    Sum      float64 // Total for category
    Count    int64   // Transaction count
    Avg      float64 // Average for category
}
```

## 🗄️ Database Schema

### transactions table

```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,              -- 'income' or 'expense'
    amount DECIMAL(15, 2) NOT NULL,         -- Transaction amount
    category VARCHAR(100) NOT NULL,         -- Category
    date TIMESTAMP NOT NULL,                -- Transaction date
    comment TEXT,                           -- Optional notes
    created_at TIMESTAMP DEFAULT NOW()      -- Creation timestamp
);

-- Performance indexes
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_category ON transactions(category);
CREATE INDEX idx_transactions_type ON transactions(type);
```

### Key SQL Queries

**Analytics with percentiles:**
```sql
SELECT
    SUM(amount) as sum,
    AVG(amount) as avg,
    COUNT(*) as count,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount) as median,
    PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount) as p90
FROM transactions
WHERE date >= $1 AND date < $2;
```

**Category aggregation:**
```sql
SELECT
    category,
    SUM(amount) as sum,
    COUNT(*) as count,
    AVG(amount) as avg
FROM transactions
WHERE date >= $1 AND date < $2
GROUP BY category
ORDER BY sum DESC;
```

## 🔄 Request Flow Example

### Create Transaction Flow

```
1. User submits form in browser
   ↓
2. JavaScript sends POST /api/items (JSON)
   ↓
3. handlers.CreateTransaction() receives request
   ├─ Parse JSON body
   ├─ Validate input format
   ├─ Call service.CreateTransaction()
   │
4. service.CreateTransaction()
   ├─ Validate business rules
   ├─ Call storage.CreateTransaction()
   │
5. storage.CreateTransaction()
   ├─ Build INSERT query with parameterized values
   ├─ Execute on PostgreSQL
   ├─ Return generated ID
   │
6. Handler returns JSON response with ID
   ↓
7. JavaScript updates UI in real-time
```

## 🛡️ Security Features

### SQL Injection Prevention
- All queries use parameterized statements
- User input never concatenated into SQL
- lib/pq driver handles escaping

Example:
```go
// ✓ Safe - parameterized
db.QueryContext(ctx, "SELECT * FROM transactions WHERE id = $1", userID)

// ✗ Unsafe - not used in this project
db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM transactions WHERE id = %d", userID))
```

### Input Validation
- Type checking (income/expense)
- Amount validation (non-negative)
- Date validation (valid and not future)
- Category required field

### CORS Configuration
- CORS middleware allows cross-origin requests
- Configurable via `corsMiddleware()` function
- All standard HTTP methods supported

## 📊 Analytics Implementation

### Median Calculation
Uses PostgreSQL `PERCENTILE_CONT` ordered set aggregate:
```sql
PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount)
```
Accurate for continuous data distribution.

### 90th Percentile
```sql
PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount)
```
Useful for identifying high-value transactions.

### Performance
- Indexes on `date`, `category`, `type` columns
- Connection pooling: 10 max open, 5 idle
- Efficient SQL with GROUP BY for aggregation

## 🎯 WBF Framework Integration

### dbpg (Database)
```go
db, err := dbpg.New(masterDSN, slaveDSNs, opts)
db.ExecContext()    // Write queries
db.QueryContext()   // Read queries
```
- Master-slave replication support
- Connection pooling
- Retry mechanisms

### zlog (Logging)
```go
zlog.Logger.Info().Str("key", value).Msg("message")
zlog.Logger.Error().Err(err).Msg("error")
```
- Structured logging
- JSON output format
- Severity levels

### config (Configuration)
```go
cfg, err := config.Load()
cfg.Database.DSN
cfg.Server.Port
```
- Load from env variables
- Support for config files
- Default values

## 🚀 Deployment Architecture

### Development
```
Developer Machine
├─ Go runtime
├─ PostgreSQL (Docker)
└─ Browser
```

### Production Scenario
```
Web Server (Linux)
├─ task_6.exe binary
├─ static/ files
└─ Configuration

Separate PostgreSQL Server
├─ PostgreSQL 15
└─ sales_tracker database
```

Connection via DSN:
```
DATABASE_DSN=postgres://user:pass@db-server:5442/sales_tracker
```

## 📈 Scalability Considerations

### Current Implementation
- Single server instance
- Connection pooling: 10 connections
- In-memory calculations only

### Future Improvements
- Database replication (read replicas via dbpg)
- Caching layer (Redis via wbf)
- Message queues (Kafka/RabbitMQ via wbf)
- Load balancing
- Microservices architecture

## 🧪 Testing Strategy

### Unit Tests (Recommended)
```
service/service_test.go    - Service layer tests
storage/postgres_test.go   - SQL query tests
models/transaction_test.go - Validation tests
handlers/handlers_test.go  - API endpoint tests
```

### Integration Tests
```
Full flow testing with real PostgreSQL
Database state verification
API response validation
```

### Manual Testing
```
test-api.ps1 - PowerShell script for testing all endpoints
Browser manual testing via UI
```

## 🔧 Configuration Flow

```
Environment Variables (highest priority)
        ↓
Config File (.env or config.yaml)
        ↓
Defaults in Code
```

**Order of precedence:**
1. `DATABASE_DSN` env variable
2. Config file
3. Default in main.go

## 📝 Code Organization Principles

1. **Separation of Concerns**
   - Handlers: HTTP only
   - Service: Business logic only
   - Storage: Database only

2. **Error Handling**
   - Errors propagate upward
   - HTTP handlers format errors
   - Logging at each layer

3. **Type Safety**
   - Strong typing throughout
   - Enum-like types for transaction types
   - Validation at domain layer

4. **Dependency Injection**
   - Storage injected into Service
   - Service injected into Handlers
   - Database injected into Storage

## 🎓 Educational Value

This project demonstrates:
- Professional Go project structure
- REST API design patterns
- Database connection management
- SQL optimization techniques
- Web framework integration
- Frontend-backend communication
- Error handling patterns
- Configuration management
- Clean code principles

---

**For detailed implementation, see source files in each module.**

