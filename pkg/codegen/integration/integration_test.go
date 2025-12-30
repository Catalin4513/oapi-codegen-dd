// Copyright 2025 DoorDash, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

//go:build integration
// +build integration

package integration

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/doordash/oapi-codegen-dd/v3/pkg/codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/specs
var specsFS embed.FS

func TestIntegration(t *testing.T) {
	specPath := os.Getenv("SPEC")

	// Collect specs to process
	specs := collectSpecs(t, specPath)
	if len(specs) == 0 {
		log.Println("No specs to process, skipping integration test")
		return
	}

	log.Printf("Found %d spec(s) to process\n", len(specs))

	cfg := codegen.Configuration{
		PackageName: "integration",
		Generate: &codegen.GenerateOptions{
			Client: true,
		},
		Client: &codegen.Client{
			Name: "IntegrationClient",
		},
	}
	cfg = cfg.Merge(codegen.NewDefaultConfiguration())

	// Track results for summary
	var (
		mu     sync.Mutex
		passed []string
		failed []string
		total  = len(specs)
	)

	for _, name := range specs {
		t.Run(fmt.Sprintf("test-%s", name), func(t *testing.T) {
			t.Parallel()

			contents, err := getFileContents(name)
			if err != nil {
				t.Fatalf("failed to download file: %s", err)
			}

			fmt.Printf("[%s] Generating code\n", name)
			res, err := codegen.Generate(contents, cfg)
			require.NoError(t, err, "failed to generate code")
			require.NotNil(t, res, "result should not be nil")

			assert.NotNil(t, res["package integration"])
			assert.NotNil(t, res["type IntegrationClient struct {"])
			assert.NotNil(t, res["RequestOptions struct {"])

			// Save generated code to a temporary directory and build it
			// Use os.MkdirTemp instead of t.TempDir() so we can control cleanup
			tmpDir, err := os.MkdirTemp("", "oapi-codegen-test-*")
			require.NoError(t, err, "failed to create temp dir")

			// Clean up temp dir after test completes (unless test fails and we want to inspect)
			defer func() {
				if !t.Failed() {
					os.RemoveAll(tmpDir)
				}
			}()

			genFile := filepath.Join(tmpDir, "generated.go")

			fmt.Printf("[%s] Saving generated code to %s\n", name, genFile)
			err = os.WriteFile(genFile, []byte(res.GetCombined()), 0644)
			require.NoError(t, err, "failed to write generated code")

			// Initialize go module
			fmt.Printf("[%s] Initializing go module\n", name)
			cmd := exec.Command("go", "mod", "init", "integration")
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("go mod init output: %s", string(output))
			}
			require.NoError(t, err, "failed to initialize go module")

			// Add replace directive to use local version of the library BEFORE go mod tidy
			fmt.Printf("[%s] Adding replace directive for local library\n", name)
			// Get the absolute path to the project root (3 levels up from this file)
			projectRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
			require.NoError(t, err, "failed to get project root path")

			cmd = exec.Command("go", "mod", "edit", "-replace", fmt.Sprintf("github.com/doordash/oapi-codegen-dd/v3=%s", projectRoot))
			cmd.Dir = tmpDir
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Logf("go mod edit output: %s", string(output))
			}
			require.NoError(t, err, "failed to add replace directive")

			// Run go mod tidy to download dependencies (after replace directive is set)
			fmt.Printf("[%s] Running go mod tidy\n", name)
			cmd = exec.Command("go", "mod", "tidy")
			cmd.Dir = tmpDir
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Logf("go mod tidy output: %s", string(output))
			}
			require.NoError(t, err, "failed to run go mod tidy")

			// Build the generated code
			fmt.Printf("[%s] Building generated code\n", name)
			cmd = exec.Command("go", "build", "-o", "/dev/null", genFile)
			cmd.Dir = tmpDir
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Logf("go build output: %s", string(output))
			}
			require.NoError(t, err, "failed to build generated code")

			fmt.Printf("[%s] Successfully built generated code\n", name)
			fmt.Printf("[%s] Generated code saved at: %s\n", name, tmpDir)

			// Track result at the end of the test
			mu.Lock()
			defer mu.Unlock()
			if t.Failed() {
				failed = append(failed, name)
			} else {
				passed = append(passed, name)
			}
		})
	}

	// Wait for all subtests to complete before printing summary
	t.Cleanup(func() {
		printSummary(total, passed, failed)
	})
}

func getFileContents(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return contents, nil
}

func collectSpecs(t *testing.T, specPath string) []string {
	var specs []string

	if specPath != "" {
		specs = append(specs, specPath)
		return specs
	}

	// Walk through testdata/specs
	err := fs.WalkDir(specsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		fileName := d.Name()
		if fileName[0] == '-' || strings.Contains(path, "/stash/") {
			return nil
		}

		if strings.HasSuffix(fileName, ".yml") || strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".json") {
			specs = append(specs, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk specs directory: %v", err)
	}

	return specs
}

func printSummary(total int, passed, failed []string) {
	log.Println("\n" + strings.Repeat("=", 80))
	log.Println("INTEGRATION TEST SUMMARY")
	log.Println(strings.Repeat("=", 80))
	log.Printf("Total specs tested: %d\n", total)
	log.Printf("✅ Passed: %d\n", len(passed))
	log.Printf("❌ Failed: %d\n", len(failed))
	log.Println(strings.Repeat("-", 80))

	if len(failed) > 0 {
		log.Println("\nFailed specs:")
		for _, spec := range failed {
			log.Printf("  ❌ %s\n", spec)
		}
	}

	if len(passed) > 0 && len(failed) == 0 {
		log.Println("\n🎉 All specs passed!")
	}

	log.Println(strings.Repeat("=", 80))
}
