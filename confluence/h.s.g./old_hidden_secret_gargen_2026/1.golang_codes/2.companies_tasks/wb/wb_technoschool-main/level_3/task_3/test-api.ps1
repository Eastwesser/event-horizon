$API_URL = "http://localhost:8080"

Write-Host "Testing Comment Tree API" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Test 1: Create root comment" -ForegroundColor Yellow
$body1 = @{
    author = "Alice"
    content = "First comment"
} | ConvertTo-Json

$comment1 = Invoke-RestMethod -Uri "$API_URL/comments" -Method Post -Body $body1 -ContentType "application/json"
$comment1 | ConvertTo-Json
$COMMENT1_ID = $comment1.id
Write-Host "Created comment ID: $COMMENT1_ID" -ForegroundColor Green
Write-Host ""

Start-Sleep -Seconds 1

Write-Host "Test 2: Create reply" -ForegroundColor Yellow
$body2 = @{
    author = "Bob"
    content = "Reply to first comment"
    parent_id = $COMMENT1_ID
} | ConvertTo-Json

$comment2 = Invoke-RestMethod -Uri "$API_URL/comments" -Method Post -Body $body2 -ContentType "application/json"
$comment2 | ConvertTo-Json
$COMMENT2_ID = $comment2.id
Write-Host ""

Start-Sleep -Seconds 1

Write-Host "Test 3: Get all comments" -ForegroundColor Yellow
$all = Invoke-RestMethod -Uri "$API_URL/comments" -Method Get
$all | ConvertTo-Json -Depth 10
Write-Host ""

Start-Sleep -Seconds 1

Write-Host "Test 4: Search comments" -ForegroundColor Yellow
$search = Invoke-RestMethod -Uri "$API_URL/comments/search?q=first" -Method Get
$search | ConvertTo-Json -Depth 10
Write-Host ""

Write-Host "All tests completed!" -ForegroundColor Green

