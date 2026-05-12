# Order Processing

Microservices backend for auth, product, and order flows, with KrakenD API Gateway, observability stack, and automated API documentation checks.

## Run Locally (Development)
```bash
docker compose up --build -d
```

Main endpoints:
- API Gateway: `http://localhost:8080`
- Swagger UI: `http://localhost:8088`
- Grafana: `http://localhost:3000`
- Prometheus: `http://localhost:9090`
- RabbitMQ Management: `http://localhost:15672`

## API Docs and Contract Automation
Swagger is generated from handler annotations:
- `service/auth/handler.go`
- `service/product/handler.go`
- `service/order/handler.go`

Day-to-day commands:
```bash
make swagger
make contract-check
make test-services
```

Run tests per service:
```bash
go test ./service/auth/...
go test ./service/product/...
go test ./service/order/...
go test ./service/notification/...
```

CI workflow (`.github/workflows/api-contract.yml`) does:
1. Regenerate Swagger specs.
2. Fail if generated specs changed unexpectedly (`git diff --exit-code docs/swagger`).
3. Validate KrakenD routes against Swagger contracts.

Swagger outputs:
- `docs/swagger/auth/swagger.json`
- `docs/swagger/product/swagger.json`
- `docs/swagger/order/swagger.json`
- `docs/swagger/swagger.json` (combined)

## Logging and Observability
Development observability stack includes:
- Prometheus + Grafana
- Loki + Promtail for log aggregation

Loki datasource is provisioned automatically via:
- `grafana/provisioning/datasources/loki.yml`

`order_service` logs are structured JSON and include fields like `trace_id`, `correlation_id`, and `event` for easier incident tracing.

## Production Deployment (Multi-VPS)
`docker-compose.yml` is intended for single-host development.

For multi-VPS production, use split compose files:
- `deploy/docker/docker-compose.prod.data.yml`
- `deploy/docker/docker-compose.prod.auth.yml`
- `deploy/docker/docker-compose.prod.product.yml`
- `deploy/docker/docker-compose.prod.order.yml`
- `deploy/docker/docker-compose.prod.gateway.yml`
- `deploy/docker/docker-compose.prod.monitoring.yml`

Example env files:
- `deploy/docker/env/.env.data.example`
- `deploy/docker/env/.env.auth.example`
- `deploy/docker/env/.env.product.example`
- `deploy/docker/env/.env.order.example`
- `deploy/docker/env/.env.gateway.example`
- `deploy/docker/env/.env.monitoring.example`

Example deploy commands:
```bash
docker compose --env-file deploy/docker/env/.env.data.example -f deploy/docker/docker-compose.prod.data.yml up --build -d
docker compose --env-file deploy/docker/env/.env.auth.example -f deploy/docker/docker-compose.prod.auth.yml up --build -d
docker compose --env-file deploy/docker/env/.env.product.example -f deploy/docker/docker-compose.prod.product.yml up --build -d
docker compose --env-file deploy/docker/env/.env.order.example -f deploy/docker/docker-compose.prod.order.yml up --build -d
docker compose --env-file deploy/docker/env/.env.gateway.example -f deploy/docker/docker-compose.prod.gateway.yml up --build -d
docker compose --env-file deploy/docker/env/.env.monitoring.example -f deploy/docker/docker-compose.prod.monitoring.yml up --build -d
```

Production notes:
- Docker internal hostnames (`auth_service`, etc.) do not work across VPS boundaries.
- Gateway upstream hosts are configurable via env vars:
  - `AUTH_SERVICE_URL`
  - `PRODUCT_SERVICE_URL`
  - `ORDER_SERVICE_URL`
- Production Prometheus scrape targets are defined in `prometheus.prod.yml`; adjust them to your real private IPs or DNS names.
