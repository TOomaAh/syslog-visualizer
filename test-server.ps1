# PowerShell test script for syslog-visualizer server

Write-Host "Syslog Visualizer Server Test" -ForegroundColor Green
Write-Host "=============================" -ForegroundColor Green
Write-Host ""

# Wait for the server to start
Start-Sleep -Seconds 2

# Function to send a UDP message
function Send-SyslogUDP {
    param([string]$Message)

    $udpClient = New-Object System.Net.Sockets.UdpClient
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($Message)
    $udpClient.Send($bytes, $bytes.Length, "localhost", 514) | Out-Null
    $udpClient.Close()
}

# Send some syslog messages via UDP
Write-Host "Sending syslog messages via UDP..." -ForegroundColor Cyan
Send-SyslogUDP "<34>Oct 11 22:14:15 server1 su[1234]: 'su root' failed for user"
Send-SyslogUDP "<13>Feb  5 17:32:18 server2 myapp: Application started successfully"
Send-SyslogUDP "<86>Dec  1 08:30:00 server3 kernel: Out of memory warning"

# Wait for messages to be processed
Start-Sleep -Seconds 1

# Query the API to verify messages were received
Write-Host ""
Write-Host "Retrieving messages via API..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/syslogs" -Method Get
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "ERROR when calling API: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "Test completed!" -ForegroundColor Green
Write-Host ""
Write-Host "To test manually:" -ForegroundColor Yellow
Write-Host "  - Health check: Invoke-RestMethod http://localhost:8080/api/health" -ForegroundColor White
Write-Host "  - View messages: Invoke-RestMethod http://localhost:8080/api/syslogs" -ForegroundColor White
