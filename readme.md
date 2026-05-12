## API Automation
Production API contract/testing guide:
- [docs/api-automation-guide.md](docs/api-automation-guide.md)

Generate Swagger docs from Go code (`swaggo/swag`):
```bash
powershell -ExecutionPolicy Bypass -File scripts/generate-swagger.ps1
```

Run KrakenD contract check against generated Swagger:
```bash
powershell -ExecutionPolicy Bypass -File scripts/test-krakend-contract.ps1
```

Run all service tests:
```bash
powershell -ExecutionPolicy Bypass -File scripts/test-services.ps1
```

Run tests for a single service:
```bash
go test ./service/auth/...
go test ./service/product/...
go test ./service/order/...
go test ./service/notification/...
```

Hosted Swagger UI:
Included in normal startup:
```bash
docker compose up --build -d
```
- URL: `http://localhost:8088`
