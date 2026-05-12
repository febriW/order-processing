$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Path "docs/swagger/auth" -Force | Out-Null
New-Item -ItemType Directory -Path "docs/swagger/product" -Force | Out-Null
New-Item -ItemType Directory -Path "docs/swagger/order" -Force | Out-Null
New-Item -ItemType Directory -Path ".gocache" -Force | Out-Null
New-Item -ItemType Directory -Path ".gomodcache" -Force | Out-Null
New-Item -ItemType Directory -Path ".gopath" -Force | Out-Null

$env:GOCACHE = (Resolve-Path ".gocache").Path
$env:GOMODCACHE = (Resolve-Path ".gomodcache").Path
$env:GOPATH = (Resolve-Path ".gopath").Path

function Invoke-SwagGenerateLocal {
    param (
        [string]$MainFile,
        [string]$ScanDirs,
        [string]$OutputDir
    )

    swag init -g $MainFile -d $ScanDirs -o $OutputDir --outputTypes json --parseDependency
    if ($LASTEXITCODE -ne 0) {
        throw "swag generation failed for $MainFile (exit code: $LASTEXITCODE)"
    }
}

function Invoke-SwagGenerateGoRun {
    param (
        [string]$MainFile,
        [string]$ScanDirs,
        [string]$OutputDir
    )

    go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g $MainFile -d $ScanDirs -o $OutputDir --outputTypes json --parseDependency
    if ($LASTEXITCODE -ne 0) {
        throw "swag generation failed for $MainFile via 'go run' (exit code: $LASTEXITCODE)"
    }
}

function Merge-SwaggerSpecs {
    param (
        [string[]]$InputFiles,
        [string]$OutputFile
    )

    if ($InputFiles.Count -eq 0) {
        throw "no swagger input files provided"
    }

    $base = Get-Content -Raw $InputFiles[0] | ConvertFrom-Json

    if ($null -eq $base.paths) {
        $base | Add-Member -NotePropertyName paths -NotePropertyValue ([ordered]@{})
    }
    if ($null -eq $base.definitions) {
        $base | Add-Member -NotePropertyName definitions -NotePropertyValue ([ordered]@{})
    }
    if ($null -eq $base.securityDefinitions) {
        $base | Add-Member -NotePropertyName securityDefinitions -NotePropertyValue ([ordered]@{})
    }
    if ($null -eq $base.tags) {
        $base | Add-Member -NotePropertyName tags -NotePropertyValue @()
    }
    if ($null -eq $base.info) {
        $base | Add-Member -NotePropertyName info -NotePropertyValue ([ordered]@{})
    }

    foreach ($input in $InputFiles | Select-Object -Skip 1) {
        $spec = Get-Content -Raw $input | ConvertFrom-Json

        if ($null -ne $spec.paths) {
            foreach ($prop in $spec.paths.PSObject.Properties) {
                $base.paths | Add-Member -NotePropertyName $prop.Name -NotePropertyValue $prop.Value -Force
            }
        }

        if ($null -ne $spec.definitions) {
            foreach ($prop in $spec.definitions.PSObject.Properties) {
                $base.definitions | Add-Member -NotePropertyName $prop.Name -NotePropertyValue $prop.Value -Force
            }
        }

        if ($null -ne $spec.securityDefinitions) {
            foreach ($prop in $spec.securityDefinitions.PSObject.Properties) {
                $base.securityDefinitions | Add-Member -NotePropertyName $prop.Name -NotePropertyValue $prop.Value -Force
            }
        }

        if ($null -ne $spec.tags) {
            $base.tags += $spec.tags
        }
    }

    $base.info | Add-Member -NotePropertyName title -NotePropertyValue "Order Processing API" -Force
    $base.info | Add-Member -NotePropertyName description -NotePropertyValue "Combined API documentation for auth, product, and order services." -Force

    $base.tags = $base.tags | Group-Object name | ForEach-Object { $_.Group[0] }

    $base | ConvertTo-Json -Depth 100 | Set-Content -Path $OutputFile -Encoding UTF8
}

$localSwag = Get-Command swag -ErrorAction SilentlyContinue
if ($null -ne $localSwag) {
    Invoke-SwagGenerateLocal "main.go" "service/auth" "docs/swagger/auth"
    Invoke-SwagGenerateLocal "main.go" "service/product" "docs/swagger/product"
    Invoke-SwagGenerateLocal "main.go" "service/order" "docs/swagger/order"
} else {
    Invoke-SwagGenerateGoRun "main.go" "service/auth" "docs/swagger/auth"
    Invoke-SwagGenerateGoRun "main.go" "service/product" "docs/swagger/product"
    Invoke-SwagGenerateGoRun "main.go" "service/order" "docs/swagger/order"
}

Merge-SwaggerSpecs -InputFiles @(
    "docs/swagger/auth/swagger.json",
    "docs/swagger/product/swagger.json",
    "docs/swagger/order/swagger.json"
) -OutputFile "docs/swagger/swagger.json"

Write-Host "Generated:"
Write-Host " - docs/swagger/auth/swagger.json"
Write-Host " - docs/swagger/product/swagger.json"
Write-Host " - docs/swagger/order/swagger.json"
Write-Host " - docs/swagger/swagger.json"
