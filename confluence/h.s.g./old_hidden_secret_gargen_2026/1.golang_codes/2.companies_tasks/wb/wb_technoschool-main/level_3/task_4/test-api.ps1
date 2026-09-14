$API_URL = "http://localhost:8080"

Write-Host "Testing Image Processor API" -ForegroundColor Cyan
Write-Host "============================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path "test-image.jpg")) {
    Write-Host "Creating test image file..." -ForegroundColor Yellow
    $width = 800
    $height = 600
    $bitmap = New-Object System.Drawing.Bitmap($width, $height)
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    $brush = [System.Drawing.Brushes]::LightBlue
    $graphics.FillRectangle($brush, 0, 0, $width, $height)
    $font = New-Object System.Drawing.Font("Arial", 24)
    $graphics.DrawString("Test Image", $font, [System.Drawing.Brushes]::Black, 300, 280)
    $bitmap.Save("test-image.jpg", [System.Drawing.Imaging.ImageFormat]::Jpeg)
    $graphics.Dispose()
    $bitmap.Dispose()
    Write-Host "Test image created: test-image.jpg" -ForegroundColor Green
    Write-Host ""
}

Write-Host "Test 1: Upload image" -ForegroundColor Yellow
try {
    $filePath = "test-image.jpg"
    $boundary = [System.Guid]::NewGuid().ToString()
    $fileBytes = [System.IO.File]::ReadAllBytes($filePath)
    $fileName = [System.IO.Path]::GetFileName($filePath)
    
    $bodyLines = @(
        "--$boundary",
        "Content-Disposition: form-data; name=`"file`"; filename=`"$fileName`"",
        "Content-Type: image/jpeg",
        "",
        [System.Text.Encoding]::GetEncoding("iso-8859-1").GetString($fileBytes),
        "--$boundary--"
    )
    
    $body = $bodyLines -join "`r`n"
    $bodyBytes = [System.Text.Encoding]::GetEncoding("iso-8859-1").GetBytes($body)
    
    $response = Invoke-RestMethod -Uri "$API_URL/upload" -Method Post -Body $bodyBytes -ContentType "multipart/form-data; boundary=$boundary"
    $response | ConvertTo-Json
    $IMAGE_ID = $response.id
    Write-Host "Image ID: $IMAGE_ID" -ForegroundColor Green
    Write-Host "Status: $($response.status)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 2

Write-Host "Test 2: Get image info" -ForegroundColor Yellow
try {
    $imageInfo = Invoke-RestMethod -Uri "$API_URL/image/$IMAGE_ID" -Method Get
    $imageInfo | ConvertTo-Json
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 3

Write-Host "Test 3: Get all images" -ForegroundColor Yellow
try {
    $allImages = Invoke-RestMethod -Uri "$API_URL/images" -Method Get
    Write-Host "Total images: $($allImages.Count)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "Test 4: Check processed image status" -ForegroundColor Yellow
try {
    $imageInfo = Invoke-RestMethod -Uri "$API_URL/image/$IMAGE_ID" -Method Get
    Write-Host "Status: $($imageInfo.status)" -ForegroundColor $(if ($imageInfo.status -eq "completed") { "Green" } else { "Yellow" })
    Write-Host ""
    
    if ($imageInfo.status -eq "completed") {
        Write-Host "Test 5: Download processed image" -ForegroundColor Yellow
        $processedUrl = "$API_URL/image/$IMAGE_ID/file?type=processed"
        Write-Host "Processed image URL: $processedUrl" -ForegroundColor Cyan
        Write-Host "Open in browser or use:" -ForegroundColor Cyan
        Write-Host "Invoke-WebRequest -Uri `"$processedUrl`" -OutFile `"processed-$IMAGE_ID.jpg`"" -ForegroundColor Gray
        Write-Host ""
        
        Write-Host "Test 6: Download thumbnail" -ForegroundColor Yellow
        $thumbUrl = "$API_URL/image/$IMAGE_ID/file?type=thumbnail"
        Write-Host "Thumbnail URL: $thumbUrl" -ForegroundColor Cyan
        Write-Host ""
        
        Write-Host "Test 7: Download original" -ForegroundColor Yellow
        $originalUrl = "$API_URL/image/$IMAGE_ID/file?type=original"
        Write-Host "Original URL: $originalUrl" -ForegroundColor Cyan
        Write-Host ""
    } else {
        Write-Host "Image is still processing. Status: $($imageInfo.status)" -ForegroundColor Yellow
        Write-Host "Wait a few seconds and check again with:" -ForegroundColor Cyan
        Write-Host "Invoke-RestMethod -Uri `"$API_URL/image/$IMAGE_ID`"" -ForegroundColor Gray
        Write-Host ""
    }
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "Test 8: Delete image" -ForegroundColor Yellow
try {
    $deleteResponse = Invoke-RestMethod -Uri "$API_URL/image/$IMAGE_ID" -Method Delete
    Write-Host "Image deleted successfully" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host "All tests completed!" -ForegroundColor Green
