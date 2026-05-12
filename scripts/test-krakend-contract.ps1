Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Path ".gocache" -Force | Out-Null
New-Item -ItemType Directory -Path ".gomodcache" -Force | Out-Null
New-Item -ItemType Directory -Path ".gopath" -Force | Out-Null

$env:GOCACHE = (Resolve-Path ".gocache").Path
$env:GOMODCACHE = (Resolve-Path ".gomodcache").Path
$env:GOPATH = (Resolve-Path ".gopath").Path

go run ./cmd/tools/contractcheck
