# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth-service ./auth-service

# Final stage - distroless static
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /app/auth-service .

USER nonroot:nonroot
EXPOSE 8083

ENTRYPOINT ["/auth-service"]
