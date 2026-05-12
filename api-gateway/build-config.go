package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type config map[string]any

func main() {
	if len(os.Args) != 3 {
		fatalf("usage: build-config <gateway-dir> <output-file>")
	}

	gatewayDir := os.Args[1]
	outputFile := os.Args[2]

	rawBase, err := os.ReadFile(filepath.Join(gatewayDir, "krakend.json"))
	if err != nil {
		fatalf("read base config: %v", err)
	}
	rawBase = []byte(os.ExpandEnv(string(rawBase)))

	var base config
	if err := json.Unmarshal(rawBase, &base); err != nil {
		fatalf("parse base config: %v", err)
	}

	filesValue, ok := base["endpoint_files"].([]any)
	if !ok {
		fatalf("endpoint_files must be an array")
	}

	endpoints := make([]any, 0)
	for _, value := range filesValue {
		fileName, ok := value.(string)
		if !ok {
			fatalf("endpoint file entries must be strings")
		}

		rawEndpoints, err := os.ReadFile(filepath.Join(gatewayDir, fileName))
		if err != nil {
			fatalf("read endpoint file %s: %v", fileName, err)
		}
		rawEndpoints = []byte(os.ExpandEnv(string(rawEndpoints)))

		var serviceEndpoints []any
		if err := json.Unmarshal(rawEndpoints, &serviceEndpoints); err != nil {
			fatalf("parse endpoint file %s: %v", fileName, err)
		}

		endpoints = append(endpoints, serviceEndpoints...)
	}

	delete(base, "endpoint_files")
	base["endpoints"] = endpoints

	rendered, err := json.MarshalIndent(base, "", "  ")
	if err != nil {
		fatalf("render config: %v", err)
	}

	if err := os.WriteFile(outputFile, append(rendered, '\n'), 0644); err != nil {
		fatalf("write output config: %v", err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
