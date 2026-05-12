package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "contract check failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	krakendFiles := []string{
		filepath.Join(root, "api-gateway", "endpoints", "auth.json"),
		filepath.Join(root, "api-gateway", "endpoints", "product.json"),
		filepath.Join(root, "api-gateway", "endpoints", "order.json"),
	}
	krakendPairs, err := getEndpointPairsFromKrakend(krakendFiles)
	if err != nil {
		return err
	}

	swaggerPairs, err := getEndpointPairsFromSwagger(filepath.Join(root, "docs", "swagger", "swagger.json"))
	if err != nil {
		return err
	}

	missingInSwagger := diffPairs(krakendPairs, swaggerPairs)
	missingInKrakend := diffPairs(swaggerPairs, krakendPairs)

	if len(missingInSwagger) > 0 || len(missingInKrakend) > 0 {
		if len(missingInSwagger) > 0 {
			fmt.Println("Endpoints in KrakenD but missing in Swagger:")
			for _, pair := range missingInSwagger {
				fmt.Printf(" - %s\n", pair)
			}
		}
		if len(missingInKrakend) > 0 {
			fmt.Println("Endpoints in Swagger but missing in KrakenD:")
			for _, pair := range missingInKrakend {
				fmt.Printf(" - %s\n", pair)
			}
		}
		return fmt.Errorf("krakend contract check failed")
	}

	fmt.Println("KrakenD contract check passed for all endpoints.")
	return nil
}

func getEndpointPairsFromKrakend(files []string) (map[string]bool, error) {
	type endpoint struct {
		Method   string `json:"method"`
		Endpoint string `json:"endpoint"`
	}

	pairs := map[string]bool{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read krakend endpoint file %s: %w", file, err)
		}
		var items []endpoint
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, fmt.Errorf("parse krakend endpoint file %s: %w", file, err)
		}
		for _, item := range items {
			key := strings.ToUpper(item.Method) + " " + item.Endpoint
			pairs[key] = true
		}
	}
	return pairs, nil
}

func getEndpointPairsFromSwagger(swaggerPath string) (map[string]bool, error) {
	type swaggerDoc struct {
		Paths map[string]map[string]any `json:"paths"`
	}

	data, err := os.ReadFile(swaggerPath)
	if err != nil {
		return nil, fmt.Errorf("read swagger spec %s: %w", swaggerPath, err)
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	var doc swaggerDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse swagger spec %s: %w", swaggerPath, err)
	}

	pairs := map[string]bool{}
	for path, methods := range doc.Paths {
		for method := range methods {
			key := strings.ToUpper(method) + " " + path
			pairs[key] = true
		}
	}
	return pairs, nil
}

func diffPairs(a, b map[string]bool) []string {
	missing := make([]string, 0)
	for key := range a {
		if !b[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}
