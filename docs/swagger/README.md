# swaggo API Docs

Swagger specs are generated from `swaggo/swag` annotations in Go source.

## Generate
```bash
make swagger
```

## Output Files
- `docs/swagger/swagger.json` (combined: auth + product + order)
- `docs/swagger/auth/swagger.json`
- `docs/swagger/product/swagger.json`
- `docs/swagger/order/swagger.json`

## Hosted Swagger UI
Included in normal startup:
```bash
docker compose up --build -d
```
- URL: `http://localhost:8088`
- Default view is `All APIs` (all endpoints together).
- Topbar dropdown also lets you switch to `Auth API`, `Product API`, and `Order API`.

## Notes
- Route/parameter docs live in:
  - `service/auth/handler.go` (inline handler annotations)
  - `service/product/handler.go` (inline handler annotations)
  - `service/order/handler.go` (inline handler annotations)
- If routes change in handlers, update these doc annotations, then regenerate specs.
