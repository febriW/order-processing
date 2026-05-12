Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-EndpointPairsFromKrakend {
    param([string]$BaseDir)

    $files = @(
        Join-Path $BaseDir "api-gateway/endpoints/auth.json"
        Join-Path $BaseDir "api-gateway/endpoints/product.json"
        Join-Path $BaseDir "api-gateway/endpoints/order.json"
    )

    $pairs = @{}
    foreach ($file in $files) {
        if (!(Test-Path -LiteralPath $file)) {
            throw "KrakenD endpoint file not found: $file"
        }
        $items = Get-Content -LiteralPath $file -Raw | ConvertFrom-Json
        foreach ($item in $items) {
            $method = [string]$item.method
            $path = [string]$item.endpoint
            $key = "{0} {1}" -f $method.ToUpperInvariant(), $path
            $pairs[$key] = $true
        }
    }
    return $pairs
}

function Get-EndpointPairsFromSwagger {
    param([string]$SwaggerPath)

    if (!(Test-Path -LiteralPath $SwaggerPath)) {
        throw "Swagger spec not found: $SwaggerPath"
    }

    $spec = Get-Content -LiteralPath $SwaggerPath -Raw | ConvertFrom-Json
    $pairs = @{}

    foreach ($pathProp in $spec.paths.PSObject.Properties) {
        $path = $pathProp.Name
        foreach ($methodProp in $pathProp.Value.PSObject.Properties) {
            $method = $methodProp.Name
            $key = "{0} {1}" -f $method.ToUpperInvariant(), $path
            $pairs[$key] = $true
        }
    }

    return $pairs
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$krakendPairs = Get-EndpointPairsFromKrakend -BaseDir $repoRoot
$swaggerPairs = Get-EndpointPairsFromSwagger -SwaggerPath (Join-Path $repoRoot "docs/swagger/swagger.json")

$missingInSwagger = @()
foreach ($pair in $krakendPairs.Keys) {
    if (-not $swaggerPairs.ContainsKey($pair)) {
        $missingInSwagger += $pair
    }
}

$missingInKrakend = @()
foreach ($pair in $swaggerPairs.Keys) {
    if (-not $krakendPairs.ContainsKey($pair)) {
        $missingInKrakend += $pair
    }
}

if ($missingInSwagger.Count -gt 0 -or $missingInKrakend.Count -gt 0) {
    if ($missingInSwagger.Count -gt 0) {
        Write-Host "Endpoints in KrakenD but missing in Swagger:"
        $missingInSwagger | Sort-Object | ForEach-Object { Write-Host " - $_" }
    }
    if ($missingInKrakend.Count -gt 0) {
        Write-Host "Endpoints in Swagger but missing in KrakenD:"
        $missingInKrakend | Sort-Object | ForEach-Object { Write-Host " - $_" }
    }
    throw "KrakenD contract check failed"
}

Write-Host "KrakenD contract check passed for all endpoints."
