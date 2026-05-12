.PHONY: swagger contract-check test-services verify-api

swagger:
	go run ./cmd/tools/swaggergen

contract-check:
	go run ./cmd/tools/contractcheck

test-services:
	go run ./cmd/tools/testservices

verify-api: swagger contract-check
