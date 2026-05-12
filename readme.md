## API Automation
Production API contract/testing guide:
- [docs/api-automation-guide.md](docs/api-automation-guide.md)

Generate Swagger docs from Go code (`swaggo/swag`):
```bash
powershell -ExecutionPolicy Bypass -File scripts/generate-swagger.ps1
```

Hosted Swagger UI:
Included in normal startup:
```bash
docker compose up --build -d
```
- URL: `http://localhost:8088`
