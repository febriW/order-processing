# Order Processing

Backend microservices untuk alur auth, product, dan order, dengan API Gateway (KrakenD), observability, dan dokumentasi API otomatis.

## Lokal (Development)
```bash
docker compose up --build -d
```

List Service:
- API Gateway: `http://localhost:8080`
- Swagger UI: `http://localhost:8088`
- Grafana: `http://localhost:3000`
- Prometheus: `http://localhost:9090`
- RabbitMQ Management: `http://localhost:15672`

## API Docs dan Contract Automation
Swagger di-generate dari annotation handler:
- `service/auth/handler.go`
- `service/product/handler.go`
- `service/order/handler.go`

List Command:
```bash
make swagger
make contract-check
make test-services
```

Untuk menjalankan Test per service:
```bash
go test ./service/auth/...
go test ./service/product/...
go test ./service/order/...
go test ./service/notification/...
```

CI workflow (`.github/workflows/api-contract.yml`) menjalankan:
1. Regenerate spec Swagger.
2. Validasi tidak ada perubahan spec tak ter-commit (`git diff --exit-code docs/swagger`).
3. Cek parity route KrakenD vs Swagger contract.

Output Swagger:
- `docs/swagger/auth/swagger.json`
- `docs/swagger/product/swagger.json`
- `docs/swagger/order/swagger.json`
- `docs/swagger/swagger.json` (gabungan)

## Logging dan Observability
Monitoring stack di development:
- Prometheus + Grafana
- Loki + Promtail untuk log aggregation

Datasource Loki diprovision otomatis via:
- `grafana/provisioning/datasources/loki.yml`

Log dari `order_service` sudah structured JSON (termasuk `trace_id`, `correlation_id`, `event`) untuk mempermudah tracing saat terjadi incident.

## Deploy
`docker-compose.yml` dipakai untuk development single-host.

Untuk production, gunakan compose terpisah masing-masing service:
- `deploy/docker/docker-compose.prod.data.yml`
- `deploy/docker/docker-compose.prod.auth.yml`
- `deploy/docker/docker-compose.prod.product.yml`
- `deploy/docker/docker-compose.prod.order.yml`
- `deploy/docker/docker-compose.prod.gateway.yml`
- `deploy/docker/docker-compose.prod.monitoring.yml`

Contoh env per layer ada di:
- `deploy/docker/env/.env.data.example`
- `deploy/docker/env/.env.auth.example`
- `deploy/docker/env/.env.product.example`
- `deploy/docker/env/.env.order.example`
- `deploy/docker/env/.env.gateway.example`
- `deploy/docker/env/.env.monitoring.example`

Contoh perintah deploy:
```bash
docker compose --env-file deploy/docker/env/.env.data.example -f deploy/docker/docker-compose.prod.data.yml up --build -d
docker compose --env-file deploy/docker/env/.env.auth.example -f deploy/docker/docker-compose.prod.auth.yml up --build -d
docker compose --env-file deploy/docker/env/.env.product.example -f deploy/docker/docker-compose.prod.product.yml up --build -d
docker compose --env-file deploy/docker/env/.env.order.example -f deploy/docker/docker-compose.prod.order.yml up --build -d
docker compose --env-file deploy/docker/env/.env.gateway.example -f deploy/docker/docker-compose.prod.gateway.yml up --build -d
docker compose --env-file deploy/docker/env/.env.monitoring.example -f deploy/docker/docker-compose.prod.monitoring.yml up --build -d
```

Catatan untuk production:
- Hostname internal docker (`auth_service`, dst) tidak bisa dipakai lintas VPS.
- Endpoint gateway bisa di-set via env:
  - `AUTH_SERVICE_URL`
  - `PRODUCT_SERVICE_URL`
  - `ORDER_SERVICE_URL`
- Scrape target Prometheus production ada di `prometheus.prod.yml`, sesuaikan dengan IP/private DNS aktual.
