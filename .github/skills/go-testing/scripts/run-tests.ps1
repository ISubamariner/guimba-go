# Run Go tests with coverage report
param(
    [string]$Package = "./...",
    [switch]$Verbose,
    [switch]$Integration
)

$ErrorActionPreference = "Stop"

Push-Location "$PSScriptRoot/../../../backend"

try {
    $args = @("test")

    if ($Verbose) { $args += "-v" }
    if ($Integration) { $args += "-tags=integration" }

    $args += "-coverprofile=coverage.out"
    $args += $Package

    Write-Host "Running: go $($args -join ' ')" -ForegroundColor Cyan
    & go @args

    if ($LASTEXITCODE -eq 0) {
        Write-Host "`nCoverage summary:" -ForegroundColor Green
        go tool cover -func=coverage.out | Select-Object -Last 1

        Write-Host "`nTo view HTML report: go tool cover -html=coverage.out" -ForegroundColor Yellow
    } else {
        Write-Host "`nTests FAILED" -ForegroundColor Red
        exit 1
    }
} finally {
    Pop-Location
}
