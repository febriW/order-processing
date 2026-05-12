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

## Production Deployment (Multi VPS)
File `docker-compose.yml` tetap dipakai untuk development (single host).

Untuk production multi VPS, gunakan file compose terpisah:
- `deploy/docker/docker-compose.prod.gateway.yml`
- `deploy/docker/docker-compose.prod.auth.yml`
- `deploy/docker/docker-compose.prod.product.yml`
- `deploy/docker/docker-compose.prod.order.yml`
- `deploy/docker/docker-compose.prod.data.yml`
- `deploy/docker/docker-compose.prod.monitoring.yml`

Contoh environment file ada di:
- `deploy/docker/env/.env.gateway.example`
- `deploy/docker/env/.env.auth.example`
- `deploy/docker/env/.env.product.example`
- `deploy/docker/env/.env.order.example`
- `deploy/docker/env/.env.data.example`
- `deploy/docker/env/.env.monitoring.example`

Contoh perintah deploy per VPS:
```bash
docker compose --env-file deploy/docker/env/.env.data.example -f deploy/docker/docker-compose.prod.data.yml up --build -d
docker compose --env-file deploy/docker/env/.env.auth.example -f deploy/docker/docker-compose.prod.auth.yml up --build -d
docker compose --env-file deploy/docker/env/.env.product.example -f deploy/docker/docker-compose.prod.product.yml up --build -d
docker compose --env-file deploy/docker/env/.env.order.example -f deploy/docker/docker-compose.prod.order.yml up --build -d
docker compose --env-file deploy/docker/env/.env.gateway.example -f deploy/docker/docker-compose.prod.gateway.yml up --build -d
docker compose --env-file deploy/docker/env/.env.monitoring.example -f deploy/docker/docker-compose.prod.monitoring.yml up --build -d
```

Catatan:
- Di production lintas VPS, hostname internal docker seperti `auth_service` tidak berlaku antar server.
- Gateway endpoint sekarang mendukung environment variable host:
  - `AUTH_SERVICE_URL`
  - `PRODUCT_SERVICE_URL`
  - `ORDER_SERVICE_URL`
- Prometheus khusus production ada di `prometheus.prod.yml`, sesuaikan target IP/private DNS service kamu.
