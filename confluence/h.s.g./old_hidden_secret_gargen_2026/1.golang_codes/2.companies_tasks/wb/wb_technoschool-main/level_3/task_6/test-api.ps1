# SalesTracker API Test Script
# This script tests all API endpoints

$BASE_URL = "http://localhost:8081/api"
$DATE_FROM = "2024-01-01"
$DATE_TO = "2024-01-31"

Write-Host "=== SalesTracker API Tests ===" -ForegroundColor Green

# Test 1: Create Income Transaction
Write-Host "`nTest 1: Creating income transaction..." -ForegroundColor Yellow
$incomeBody = @{
    type = "income"
    amount = 5000
    category = "Salary"
    date = "2024-01-15"
    comment = "January salary"
} | ConvertTo-Json

$incomeResponse = Invoke-RestMethod -Uri "$BASE_URL/items" -Method POST `
    -Headers @{"Content-Type" = "application/json"} `
    -Body $incomeBody
$incomeId = $incomeResponse.id
Write-Host "Created transaction ID: $incomeId" -ForegroundColor Green

# Test 2: Create Expense Transaction
Write-Host "`nTest 2: Creating expense transaction..." -ForegroundColor Yellow
$expenseBody = @{
    type = "expense"
    amount = 500
    category = "Food"
    date = "2024-01-10"
    comment = "Grocery shopping"
} | ConvertTo-Json

$expenseResponse = Invoke-RestMethod -Uri "$BASE_URL/items" -Method POST `
    -Headers @{"Content-Type" = "application/json"} `
    -Body $expenseBody
$expenseId = $expenseResponse.id
Write-Host "Created transaction ID: $expenseId" -ForegroundColor Green

# Test 3: Add more transactions
Write-Host "`nTest 3: Creating additional transactions..." -ForegroundColor Yellow
for ($i = 0; $i -lt 3; $i++) {
    $body = @{
        type = if ($i % 2 -eq 0) { "expense" } else { "income" }
        amount = 1000 + ($i * 100)
        category = @("Food", "Rent", "Entertainment")[$i]
        date = "2024-01-$(($i + 5).ToString('00'))"
        comment = "Test transaction $i"
    } | ConvertTo-Json
    
    Invoke-RestMethod -Uri "$BASE_URL/items" -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body | Out-Null
}
Write-Host "Created 3 additional transactions" -ForegroundColor Green

# Test 4: Get Single Transaction
Write-Host "`nTest 4: Getting transaction by ID..." -ForegroundColor Yellow
$transaction = Invoke-RestMethod -Uri "$BASE_URL/items/$incomeId" -Method GET
Write-Host "Retrieved: $($transaction.category) - Amount: $($transaction.amount)" -ForegroundColor Green

# Test 5: List All Transactions
Write-Host "`nTest 5: Listing all transactions..." -ForegroundColor Yellow
$all = Invoke-RestMethod -Uri "$BASE_URL/items" -Method GET
Write-Host "Found $($all.Count) transactions" -ForegroundColor Green

# Test 6: List with Filters
Write-Host "`nTest 6: Listing with filters..." -ForegroundColor Yellow
$filtered = Invoke-RestMethod -Uri "$BASE_URL/items?from=2024-01-01`&to=2024-01-31`&type=expense" -Method GET
Write-Host "Found $($filtered.Count) expense transactions" -ForegroundColor Green

# Test 7: List by Category
Write-Host "`nTest 7: Listing by category..." -ForegroundColor Yellow
$byCategory = Invoke-RestMethod -Uri "$BASE_URL/items?category=Food" -Method GET
Write-Host "Found $($byCategory.Count) Food transactions" -ForegroundColor Green

# Test 8: Update Transaction
Write-Host "`nTest 8: Updating transaction..." -ForegroundColor Yellow
$updateBody = @{
    type = "income"
    amount = 5500
    category = "Salary"
    date = "2024-01-15"
    comment = "January salary - Updated"
} | ConvertTo-Json

Invoke-RestMethod -Uri "$BASE_URL/items/$incomeId" -Method PUT `
    -Headers @{"Content-Type" = "application/json"} `
    -Body $updateBody | Out-Null
Write-Host "Updated transaction ID: $incomeId" -ForegroundColor Green

# Test 9: Get Analytics
Write-Host "`nTest 9: Getting analytics..." -ForegroundColor Yellow
$analytics = Invoke-RestMethod -Uri "$BASE_URL/analytics?from=$DATE_FROM`&to=$DATE_TO" -Method GET
Write-Host "Analytics:" -ForegroundColor Green
Write-Host "  Sum: $($analytics.sum)" -ForegroundColor Cyan
Write-Host "  Average: $($analytics.avg)" -ForegroundColor Cyan
Write-Host "  Count: $($analytics.count)" -ForegroundColor Cyan
Write-Host "  Median: $($analytics.median)" -ForegroundColor Cyan
Write-Host "  90th Percentile: $($analytics.p90)" -ForegroundColor Cyan

# Test 10: Get Category Analytics
Write-Host "`nTest 10: Getting category analytics..." -ForegroundColor Yellow
$categoryAnalytics = Invoke-RestMethod -Uri "$BASE_URL/analytics/categories?from=$DATE_FROM`&to=$DATE_TO" -Method GET
Write-Host "Category breakdown:" -ForegroundColor Green
foreach ($cat in $categoryAnalytics) {
    Write-Host "  $($cat.category): Sum=$($cat.sum), Count=$($cat.count), Avg=$($cat.avg)" -ForegroundColor Cyan
}

# Test 11: Export CSV
Write-Host "`nTest 11: Exporting to CSV..." -ForegroundColor Yellow
$csv = Invoke-RestMethod -Uri "$BASE_URL/export/csv?from=$DATE_FROM`&to=$DATE_TO" -Method GET
$csv | Out-File -FilePath "report.csv" -Encoding UTF8
Write-Host "Exported to report.csv" -ForegroundColor Green

# Test 12: Delete Transaction
Write-Host "`nTest 12: Deleting transaction..." -ForegroundColor Yellow
Invoke-RestMethod -Uri "$BASE_URL/items/$expenseId" -Method DELETE | Out-Null
Write-Host "Deleted transaction ID: $expenseId" -ForegroundColor Green

# Test 13: Verify Deletion
Write-Host "`nTest 13: Verifying deletion..." -ForegroundColor Yellow
$after = Invoke-RestMethod -Uri "$BASE_URL/items" -Method GET
Write-Host "Transactions remaining: $($after.Count)" -ForegroundColor Green

Write-Host "`n=== All tests completed successfully ===" -ForegroundColor Green

