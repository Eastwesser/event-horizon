# PowerShell script for testing DelayedNotifier API

$API_URL = "http://localhost:8080"

Write-Host "Testing DelayedNotifier API" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Calculate time 1 minute in the future
$SCHEDULED_TIME = (Get-Date).AddMinutes(1).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

Write-Host "Test 1: Creating a console notification" -ForegroundColor Yellow
Write-Host "-------------------------------------------" -ForegroundColor Yellow

$body = @{
    channel = "console"
    recipient = "test-user"
    subject = "Test Notification"
    message = "This is a test notification from API test script"
    scheduled_at = $SCHEDULED_TIME
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$API_URL/notify" -Method Post -Body $body -ContentType "application/json"
$response | ConvertTo-Json
$NOTIFICATION_ID = $response.id
Write-Host ""
Write-Host "Created notification with ID: $NOTIFICATION_ID" -ForegroundColor Green
Write-Host ""

Start-Sleep -Seconds 2

Write-Host "Test 2: Getting notification status" -ForegroundColor Yellow
Write-Host "---------------------------------------" -ForegroundColor Yellow
$status = Invoke-RestMethod -Uri "$API_URL/notify/$NOTIFICATION_ID" -Method Get
$status | ConvertTo-Json
Write-Host ""

Start-Sleep -Seconds 2

Write-Host "Test 3: Getting all notifications" -ForegroundColor Yellow
Write-Host "-------------------------------------" -ForegroundColor Yellow
$all = Invoke-RestMethod -Uri "$API_URL/notify" -Method Get
$all | ConvertTo-Json
Write-Host ""

Start-Sleep -Seconds 2

Write-Host "Test 4: Creating and cancelling a notification" -ForegroundColor Yellow
Write-Host "---------------------------------------------------" -ForegroundColor Yellow
$SCHEDULED_TIME_FUTURE = (Get-Date).AddMinutes(5).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$body2 = @{
    channel = "console"
    recipient = "cancel-test"
    message = "This notification will be cancelled"
    scheduled_at = $SCHEDULED_TIME_FUTURE
} | ConvertTo-Json

$response2 = Invoke-RestMethod -Uri "$API_URL/notify" -Method Post -Body $body2 -ContentType "application/json"
$NOTIFICATION_ID_2 = $response2.id
Write-Host "Created notification: $NOTIFICATION_ID_2"

Start-Sleep -Seconds 2

Write-Host "Cancelling notification..."
$cancel = Invoke-RestMethod -Uri "$API_URL/notify/$NOTIFICATION_ID_2" -Method Delete
$cancel | ConvertTo-Json
Write-Host ""

Start-Sleep -Seconds 2

Write-Host "Checking status after cancellation:"
$cancelled = Invoke-RestMethod -Uri "$API_URL/notify/$NOTIFICATION_ID_2" -Method Get
$cancelled | ConvertTo-Json
Write-Host ""

Write-Host "All tests completed!" -ForegroundColor Green
Write-Host ""
Write-Host "Tip: The first notification should be sent in ~1 minute." -ForegroundColor Cyan
Write-Host "    Check the server console output for the notification." -ForegroundColor Cyan
