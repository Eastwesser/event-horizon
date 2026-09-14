$API_URL = "http://localhost:8080"

Write-Host "Testing URL Shortener API" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Test 1: Shorten URL" -ForegroundColor Yellow
$body = @{
    url = "https://www.google.com"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$API_URL/shorten" -Method Post -Body $body -ContentType "application/json"
$response | ConvertTo-Json
$SHORT_URL = $response.short_url
Write-Host "Short URL: $SHORT_URL" -ForegroundColor Green
Write-Host ""

Start-Sleep -Seconds 1

Write-Host "Test 2: Shorten with custom name" -ForegroundColor Yellow
$body2 = @{
    url = "https://www.github.com"
    custom = "gh"
} | ConvertTo-Json

$response2 = Invoke-RestMethod -Uri "$API_URL/shorten" -Method Post -Body $body2 -ContentType "application/json"
$response2 | ConvertTo-Json
Write-Host ""

Start-Sleep -Seconds 1

Write-Host "Test 3: Get analytics" -ForegroundColor Yellow
Write-Host "Visit the short URL in browser: $API_URL/s/$SHORT_URL" -ForegroundColor Cyan
Write-Host "Then check analytics..."
Start-Sleep -Seconds 2

$analytics = Invoke-RestMethod -Uri "$API_URL/analytics/$SHORT_URL" -Method Get
$analytics | ConvertTo-Json -Depth 10
Write-Host ""

Write-Host "All tests completed!" -ForegroundColor Green

