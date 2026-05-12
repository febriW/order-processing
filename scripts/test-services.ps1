Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$serviceRoot = Join-Path $repoRoot "service"

if (!(Test-Path -LiteralPath $serviceRoot -PathType Container)) {
    throw "service directory not found: $serviceRoot"
}

$serviceDirs = Get-ChildItem -LiteralPath $serviceRoot -Directory | Sort-Object Name
if ($serviceDirs.Count -eq 0) {
    throw "no services found under: $serviceRoot"
}

foreach ($svc in $serviceDirs) {
    $pkg = "./service/$($svc.Name)/..."
    $hasGoFiles = (Get-ChildItem -LiteralPath $svc.FullName -Recurse -File -Filter "*.go" | Measure-Object).Count -gt 0
    if (-not $hasGoFiles) {
        Write-Host "Skipping $pkg (no Go packages found)"
        continue
    }
    Write-Host "Running tests for $pkg"
    go test $pkg
}

Write-Host "All service tests passed."
