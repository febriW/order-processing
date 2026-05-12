package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

type swaggerSpec struct {
	Swagger             string                    `json:"swagger,omitempty"`
	Info                map[string]any            `json:"info,omitempty"`
	Paths               map[string]map[string]any `json:"paths,omitempty"`
	Definitions         map[string]any            `json:"definitions,omitempty"`
	SecurityDefinitions map[string]any            `json:"securityDefinitions,omitempty"`
	Tags                []map[string]any          `json:"tags,omitempty"`
	Other               map[string]any            `json:"-"`
}

func (s *swaggerSpec) UnmarshalJSON(data []byte) error {
	type alias swaggerSpec
	a := alias{}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*s = swaggerSpec(a)

	all := map[string]any{}
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	delete(all, "swagger")
	delete(all, "info")
	delete(all, "paths")
	delete(all, "definitions")
	delete(all, "securityDefinitions")
	delete(all, "tags")
	s.Other = all
	return nil
}

func (s swaggerSpec) MarshalJSON() ([]byte, error) {
	all := map[string]any{}
	for k, v := range s.Other {
		all[k] = v
	}
	if s.Swagger != "" {
		all["swagger"] = s.Swagger
	}
	if s.Info != nil {
		all["info"] = s.Info
	}
	if s.Paths != nil {
		all["paths"] = s.Paths
	}
	if s.Definitions != nil {
		all["definitions"] = s.Definitions
	}
	if s.SecurityDefinitions != nil {
		all["securityDefinitions"] = s.SecurityDefinitions
	}
	if s.Tags != nil {
		all["tags"] = s.Tags
	}
	return json.Marshal(all)
}

type serviceSpec struct {
	name    string
	scanDir string
	outDir  string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "swagger generation failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := ensureDirs(root); err != nil {
		return err
	}

	services := []serviceSpec{
		{name: "auth", scanDir: filepath.Join("service", "auth"), outDir: filepath.Join("docs", "swagger", "auth")},
		{name: "product", scanDir: filepath.Join("service", "product"), outDir: filepath.Join("docs", "swagger", "product")},
		{name: "order", scanDir: filepath.Join("service", "order"), outDir: filepath.Join("docs", "swagger", "order")},
	}

	for _, svc := range services {
		if err := generateSwagger(root, svc); err != nil {
			return err
		}
	}

	inputs := []string{
		filepath.Join("docs", "swagger", "auth", "swagger.json"),
		filepath.Join("docs", "swagger", "product", "swagger.json"),
		filepath.Join("docs", "swagger", "order", "swagger.json"),
	}

	if err := mergeSwaggerSpecs(root, inputs, filepath.Join("docs", "swagger", "swagger.json")); err != nil {
		return err
	}

	fmt.Println("Generated:")
	fmt.Println(" - docs/swagger/auth/swagger.json")
	fmt.Println(" - docs/swagger/product/swagger.json")
	fmt.Println(" - docs/swagger/order/swagger.json")
	fmt.Println(" - docs/swagger/swagger.json")
	return nil
}

func ensureDirs(root string) error {
	dirs := []string{
		filepath.Join("docs", "swagger", "auth"),
		filepath.Join("docs", "swagger", "product"),
		filepath.Join("docs", "swagger", "order"),
		".gocache",
		".gomodcache",
		".gopath",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func generateSwagger(root string, svc serviceSpec) error {
	env := os.Environ()
	env = append(env,
		"GOCACHE="+filepath.Join(root, ".gocache"),
		"GOMODCACHE="+filepath.Join(root, ".gomodcache"),
		"GOPATH="+filepath.Join(root, ".gopath"),
	)

	args := []string{"run", "github.com/swaggo/swag/cmd/swag@v1.16.4", "init", "-g", "main.go", "-d", svc.scanDir, "-o", svc.outDir, "--outputTypes", "json", "--parseDependency"}
	if _, err := exec.LookPath("swag"); err == nil {
		args = []string{"tool", "-modfile=go.mod"}
		_ = args
		cmd := exec.Command("swag", "init", "-g", "main.go", "-d", svc.scanDir, "-o", svc.outDir, "--outputTypes", "json", "--parseDependency")
		cmd.Dir = root
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("swag generation failed for %s via local swag: %w", svc.name, err)
		}
		return nil
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("swag generation failed for %s via go run: %w", svc.name, err)
	}
	return nil
}

func mergeSwaggerSpecs(root string, inputFiles []string, outputFile string) error {
	if len(inputFiles) == 0 {
		return errors.New("no swagger input files provided")
	}

	base, err := readSwagger(filepath.Join(root, inputFiles[0]))
	if err != nil {
		return err
	}
	ensureSpecMaps(&base)

	for _, input := range inputFiles[1:] {
		spec, err := readSwagger(filepath.Join(root, input))
		if err != nil {
			return err
		}
		ensureSpecMaps(&spec)

		for k, v := range spec.Paths {
			base.Paths[k] = v
		}
		for k, v := range spec.Definitions {
			base.Definitions[k] = v
		}
		for k, v := range spec.SecurityDefinitions {
			base.SecurityDefinitions[k] = v
		}
		base.Tags = append(base.Tags, spec.Tags...)
	}

	if base.Info == nil {
		base.Info = map[string]any{}
	}
	base.Info["title"] = "Order Processing API"
	base.Info["description"] = "Combined API documentation for auth, product, and order services."

	base.Tags = dedupeTags(base.Tags)

	jsonBytes, err := marshalStableJSON(base)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, outputFile), jsonBytes, 0o644)
}

func ensureSpecMaps(spec *swaggerSpec) {
	if spec.Paths == nil {
		spec.Paths = map[string]map[string]any{}
	}
	if spec.Definitions == nil {
		spec.Definitions = map[string]any{}
	}
	if spec.SecurityDefinitions == nil {
		spec.SecurityDefinitions = map[string]any{}
	}
	if spec.Info == nil {
		spec.Info = map[string]any{}
	}
	if spec.Other == nil {
		spec.Other = map[string]any{}
	}
}

func readSwagger(path string) (swaggerSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return swaggerSpec{}, fmt.Errorf("read swagger file %s: %w", path, err)
	}
	var spec swaggerSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return swaggerSpec{}, fmt.Errorf("parse swagger file %s: %w", path, err)
	}
	return spec, nil
}

func dedupeTags(tags []map[string]any) []map[string]any {
	seen := map[string]bool{}
	result := make([]map[string]any, 0, len(tags))
	for _, tag := range tags {
		name, _ := tag["name"].(string)
		if name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, tag)
	}
	sort.Slice(result, func(i, j int) bool {
		return fmt.Sprint(result[i]["name"]) < fmt.Sprint(result[j]["name"])
	})
	return result
}

func marshalStableJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "    "); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}
