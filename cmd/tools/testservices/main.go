package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "service tests failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	serviceRoot := filepath.Join(root, "service")
	entries, err := os.ReadDir(serviceRoot)
	if err != nil {
		return fmt.Errorf("read service directory: %w", err)
	}

	services := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			services = append(services, entry.Name())
		}
	}
	if len(services) == 0 {
		return fmt.Errorf("no services found under %s", serviceRoot)
	}
	sort.Strings(services)

	for _, svc := range services {
		svcPath := filepath.Join(serviceRoot, svc)
		hasGoFiles, err := containsGoFiles(svcPath)
		if err != nil {
			return err
		}

		pkg := "./service/" + svc + "/..."
		if !hasGoFiles {
			fmt.Printf("Skipping %s (no Go packages found)\n", pkg)
			continue
		}

		fmt.Printf("Running tests for %s\n", pkg)
		cmd := exec.Command("go", "test", pkg)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("tests failed for %s: %w", pkg, err)
		}
	}

	fmt.Println("All service tests passed.")
	return nil
}

func containsGoFiles(dir string) (bool, error) {
	found := false
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(d.Name()) == ".go" {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("scan %s: %w", dir, err)
	}
	return found, nil
}
