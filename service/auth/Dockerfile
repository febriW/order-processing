# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

# copy all folder service (auth, product, order, etc)
COPY . .

# building
WORKDIR /app/service/auth
RUN go build -o auth-service .

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /app/service/auth/auth-service .

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/auth-service"]
