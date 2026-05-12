# API Automation Guide (swaggo)

## Goal
Generate API documentation from Go code using `swaggo/swag`, then use generated specs for contract checks.

## Best Fit For This Project
Use `swaggo` annotations in each service and generate specs automatically:

1. Source of truth: route + request/response annotations in Go files.
2. Generated output: Swagger JSON per service under `docs/swagger`.
3. CI gate: regenerate + contract test against KrakenD.

## Implemented Setup
- Annotation files:
  - `service/auth/swag_docs.go`
  - `service/product/swag_docs.go`
  - `service/order/swag_docs.go`
- Generator script:
  - `scripts/generate-swagger.ps1`
- Generated artifacts:
  - `docs/swagger/auth/swagger.json`
  - `docs/swagger/product/swagger.json`
  - `docs/swagger/order/swagger.json`

## Generate Specs
```bash
powershell -ExecutionPolicy Bypass -File scripts/generate-swagger.ps1
```

## Hosted Swagger UI
Included in normal startup:
```bash
docker compose up --build -d
```
- Open `http://localhost:8088`

## Automation Flow (Production)
1. Regenerate Swagger specs in CI.
2. Fail CI if generated specs changed unexpectedly (enforce committed docs).
3. Run contract/smoke tests against KrakenD (`http://localhost:8080`).

## Notes
- Swagger UI is available in the main `docker compose` stack at `http://localhost:8088`.
- If handler routes or payload structs change, update swagger annotations first, then regenerate specs.
