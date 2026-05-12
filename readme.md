## API Automation
Generate API documentation from Go code (`swaggo/swag`) and verify KrakenD contract parity in CI.

Implemented setup:
- Annotation files:
  - `service/auth/handler.go` (inline handler annotations)
  - `service/product/handler.go` (inline handler annotations)
  - `service/order/handler.go` (inline handler annotations)
- Tool commands:
  - `cmd/tools/swaggergen/main.go`
  - `cmd/tools/contractcheck/main.go`
  - `cmd/tools/testservices/main.go`
- Wrapper:
  - `Makefile`
- CI workflow:
  - `.github/workflows/api-contract.yml`
- Generated artifacts:
  - `docs/swagger/auth/swagger.json`
  - `docs/swagger/product/swagger.json`
  - `docs/swagger/order/swagger.json`
  - `docs/swagger/swagger.json` (combined)

Generate Swagger docs from Go code (`swaggo/swag`):
```bash
make swagger
```

Run KrakenD contract check against generated Swagger:
```bash
make contract-check
```

Run all service tests:
```bash
make test-services
```

Run tests for a single service:
```bash
go test ./service/auth/...
go test ./service/product/...
go test ./service/order/...
go test ./service/notification/...
```

Automation flow in CI:
1. Regenerate Swagger specs.
2. Fail CI if generated specs changed unexpectedly (`git diff --exit-code docs/swagger`).
3. Run KrakenD route contract check against generated Swagger.

Hosted Swagger UI:
Included in normal startup:
```bash
docker compose up --build -d
```
- URL: `http://localhost:8088`
- If handler routes or payload structs change, update swagger annotations first, then regenerate specs.
