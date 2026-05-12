FROM golang:1.25.3-alpine AS config_builder

WORKDIR /src
COPY krakend.json ./krakend.json
COPY endpoints ./endpoints
COPY build-config.go /build-config.go
RUN go run /build-config.go /src /tmp/krakend.json

FROM devopsfaith/krakend:2.4.1

COPY --from=config_builder /tmp/krakend.json /etc/krakend/krakend.json
COPY certs /certs
