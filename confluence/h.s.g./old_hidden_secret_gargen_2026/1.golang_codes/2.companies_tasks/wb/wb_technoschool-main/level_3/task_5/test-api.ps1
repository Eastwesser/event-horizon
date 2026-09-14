$API_URL = "http://localhost:8080"

Write-Host "Testing EventBooker API" -ForegroundColor Cyan
Write-Host "=======================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Test 1: Create event" -ForegroundColor Yellow
try {
    $event = @{
        name = "Tech Conference 2025"
        description = "Annual technology conference"
        event_date = (Get-Date).AddDays(30).ToString("yyyy-MM-ddTHH:mm:ssZ")
        total_seats = 100
        booking_timeout_min = 15
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$API_URL/events" -Method Post -Body $event -ContentType "application/json"
    $EVENT_ID = $response.id
    Write-Host "Event created successfully!" -ForegroundColor Green
    Write-Host "Event ID: $EVENT_ID" -ForegroundColor Green
    Write-Host "Available seats: $($response.available_seats)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 1

Write-Host "Test 2: Get all events" -ForegroundColor Yellow
try {
    $events = Invoke-RestMethod -Uri "$API_URL/events" -Method Get
    Write-Host "Total events: $($events.Count)" -ForegroundColor Green
    $events | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "Test 3: Book a seat" -ForegroundColor Yellow
try {
    $booking = @{
        user_email = "user@example.com"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$API_URL/events/$EVENT_ID/book" -Method Post -Body $booking -ContentType "application/json"
    $BOOKING_ID = $response.id
    Write-Host "Booking created successfully!" -ForegroundColor Green
    Write-Host "Booking ID: $BOOKING_ID" -ForegroundColor Green
    Write-Host "Status: $($response.status)" -ForegroundColor Green
    Write-Host "Expires at: $($response.expires_at)" -ForegroundColor Yellow
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

Write-Host "Test 4: Get event info (after booking)" -ForegroundColor Yellow
try {
    $event = Invoke-RestMethod -Uri "$API_URL/events/$EVENT_ID" -Method Get
    Write-Host "Available seats: $($event.available_seats) / $($event.total_seats)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "Test 5: Confirm booking" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$API_URL/bookings/$BOOKING_ID/confirm" -Method Post
    Write-Host "Booking confirmed successfully!" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

Write-Host "Test 6: Get user bookings" -ForegroundColor Yellow
try {
    $bookings = Invoke-RestMethod -Uri "$API_URL/bookings?email=user@example.com" -Method Get
    Write-Host "Total bookings: $($bookings.Count)" -ForegroundColor Green
    $bookings | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "Test 7: Book another seat (should expire)" -ForegroundColor Yellow
try {
    $booking = @{
        user_email = "test@example.com"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$API_URL/events/$EVENT_ID/book" -Method Post -Body $booking -ContentType "application/json"
    $TEMP_BOOKING_ID = $response.id
    Write-Host "Temporary booking created: $TEMP_BOOKING_ID" -ForegroundColor Green
    Write-Host "This booking will expire in $($response.expires_at)" -ForegroundColor Yellow
    Write-Host "Wait 30+ seconds to see it auto-cancelled by scheduler" -ForegroundColor Cyan
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "Test 8: Get event bookings" -ForegroundColor Yellow
try {
    $bookings = Invoke-RestMethod -Uri "$API_URL/events/$EVENT_ID/bookings" -Method Get
    Write-Host "Total bookings for event: $($bookings.Count)" -ForegroundColor Green
    $bookings | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "All tests completed!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Open http://localhost:8080 - User interface" -ForegroundColor White
Write-Host "2. Open http://localhost:8080/admin - Admin panel" -ForegroundColor White
Write-Host "3. Watch scheduler cancel unconfirmed bookings after 30 seconds" -ForegroundColor White


