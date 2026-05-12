# Build stage
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/service/product
RUN go build -trimpath -ldflags="-s -w" -o product-service .

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /app/service/product/product-service .

USER nonroot:nonroot

EXPOSE 8082

ENTRYPOINT ["/product-service"]
