# SalesTracker - Installation & Running Guide

## ⚙️ System Requirements

- **Go**: 1.21 or higher
- **Docker & Docker Compose**: For PostgreSQL
- **PostgreSQL**: 15 (runs in Docker)
- **Browser**: Modern browser with JavaScript enabled

## 📦 Database Setup

### Option 1: Using Docker (Recommended)

**Start PostgreSQL:**
```bash
make docker-up
```

Or manually:
```bash
docker-compose up -d
```

**Database Credentials:**
- **Host:** localhost
- **Port:** 5442
- **User:** postgres
- **Password:** 8832
- **Database:** sales_tracker

**Verify PostgreSQL is running:**
```bash
docker-compose logs postgres
```

**Stop PostgreSQL:**
```bash
make docker-down
```

### Option 2: Manual PostgreSQL Installation

If you have PostgreSQL 15 installed locally, update the connection string in `env.example` or set the `DATABASE_DSN` environment variable:

```bash
export DATABASE_DSN="postgres://postgres:8832@localhost:5442/sales_tracker?sslmode=disable"
```

## 🚀 Running the Application

### Method 1: Direct Go Run (Requires Go installed)

```bash
# Download dependencies
go mod download

# Run the application
go run main.go
```

### Method 2: Using Pre-built Binary

```bash
# Build the binary (if not already built)
go build -o task_6.exe main.go

# Run the binary
./task_6.exe
```

On Windows:
```cmd
task_6.exe
```

On Linux/macOS:
```bash
./task_6.exe
```

### Method 3: Using Make

```bash
# Build and run with Docker PostgreSQL
make run

# Just build
make build

# Clean build artifacts
make clean
```

## 🌐 Accessing the Application

After starting the server, open your browser and navigate to:

```
http://localhost:8080
```

You should see the SalesTracker dashboard with:
- ➕ Add Transaction form
- 🔍 Filters
- 📋 Transactions table
- 📊 Analytics dashboard

## 📝 Environment Variables

You can customize the application behavior using environment variables:

```bash
# Database connection
DATABASE_DSN="postgres://postgres:8832@localhost:5442/sales_tracker?sslmode=disable"

# Server configuration
SERVER_PORT=8080          # Default: 8080
SERVER_HOST=0.0.0.0      # Default: localhost
```

**Load from file:**
Copy `env.example` to `.env` (optional - app will use defaults):

```bash
cp env.example .env
```

## 🧪 Testing the API

### Using PowerShell Test Script

```bash
# On Windows
.\test-api.ps1
```

The script will:
- Create test transactions
- List and filter transactions
- Get analytics
- Export to CSV
- Test CRUD operations

### Using cURL

**Create a transaction:**
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

**List transactions:**
```bash
curl "http://localhost:8080/api/items?from=2024-01-01&to=2024-01-31"
```

**Get analytics:**
```bash
curl "http://localhost:8080/api/analytics?from=2024-01-01&to=2024-01-31"
```

## 🐛 Troubleshooting

### Port 8080 Already in Use

**Solution:** Change the port:
```bash
SERVER_PORT=9000 go run main.go
```

### Database Connection Failed

**Error:** `Failed to connect to database`

**Solutions:**
1. Verify PostgreSQL is running:
   ```bash
   docker-compose logs postgres
   ```

2. Check connection string in `env.example` or `DATABASE_DSN` variable

3. Ensure port 5442 is accessible:
   ```bash
   # Windows PowerShell
   Test-NetConnection -ComputerName localhost -Port 5442
   
   # Linux/macOS
   nc -zv localhost 5442
   ```

4. Verify credentials (default: user `postgres`, password `8832`)

### Cannot Connect to Browser

1. Wait 3-5 seconds for the server to fully start
2. Check console output for startup messages
3. Verify the port is correct (default: 8080)
4. Try: http://127.0.0.1:8080

### Compilation Errors

**Error:** `unknown revision`

**Solution:** Ensure go.mod dependencies are correct:
```bash
go mod tidy
go mod verify
```

**Error:** `module not found`

**Solution:**
```bash
go mod download
go mod tidy
```

## 📋 Makefile Commands

```bash
make help          # Show available commands
make docker-up     # Start PostgreSQL
make docker-down   # Stop PostgreSQL
make docker-logs   # View PostgreSQL logs
make run           # Run application (starts DB first)
make build         # Build executable
make test          # Run tests
make clean         # Remove build artifacts
```

## 🎯 Typical Workflow

### First Time Setup

```bash
# 1. Start database
make docker-up

# 2. Wait for database to be ready (about 10 seconds)
# Check with:
docker-compose logs postgres

# 3. Run application
go run main.go

# 4. Open browser
# Navigate to http://localhost:8080
```

### During Development

```bash
# Terminal 1: Keep database running
docker-compose logs -f postgres

# Terminal 2: Run application (auto-recompiles on changes)
go run main.go
```

### Deployment Build

```bash
# 1. Build binary
make build

# 2. Create deployment archive
# - task_6.exe
# - static/index.html
# - docker-compose.yml (for reference)

# 3. On target machine:
# - Start database: docker-compose up -d
# - Run binary: ./task_6.exe
```

## ✅ Verification Steps

After starting the application, verify it's working:

1. **Browser test:**
   - Open http://localhost:8080
   - Add a test transaction
   - Verify it appears in the table

2. **API health check:**
   ```bash
   curl http://localhost:8080/health
   ```
   Expected response: `{"status":"ok"}`

3. **Database connection:**
   - Transactions appear in the UI
   - Analytics calculations work
   - CSV export completes successfully

## 📚 Additional Resources

- **README.md** - Project overview and features
- **QUICKSTART.md** - Quick reference guide
- **API Documentation** - Available at `/api/*` endpoints

## 🆘 Getting Help

1. Check the troubleshooting section above
2. Review console logs for error messages
3. Check if PostgreSQL is running:
   ```bash
   docker ps
   ```

4. Verify the database exists:
   ```bash
   docker-compose exec postgres psql -U postgres -l
   ```

## 🔄 Restarting from Scratch

If you need to completely reset:

```bash
# Stop and remove containers
docker-compose down -v

# Remove build artifacts
make clean

# Start fresh
make docker-up
sleep 3
go run main.go
```

---

**Happy tracking! 💰**

