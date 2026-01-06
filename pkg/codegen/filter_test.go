// Copyright 2025 DoorDash, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package codegen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterOperationsByTag(t *testing.T) {
	packageName := "testswagger"
	t.Run("include tags", func(t *testing.T) {
		cfg := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Include: FilterParamsConfig{
					Tags: []string{"hippo", "giraffe", "cat"},
				},
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		// Run our code generation:
		code, err := Generate([]byte(testDocument), cfg)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		assert.Contains(t, code.GetCombined(), `type CatDeadCause string`)
	})

	t.Run("exclude tags", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Exclude: FilterParamsConfig{
					Tags: []string{"hippo", "giraffe", "cat"},
				},
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		// Run our code generation:
		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.NotContains(t, code.GetCombined(), `"/cat"`)
	})
}

func TestFilterOperationsByOperationID(t *testing.T) {
	packageName := "testswagger"

	t.Run("include operation ids", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Include: FilterParamsConfig{
					OperationIDs: []string{"getCatStatus"},
				},
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		// Run our code generation:
		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Contains(t, code.GetCombined(), `type CatDeadCause string`)
	})

	t.Run("exclude operation ids", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Exclude: FilterParamsConfig{
					OperationIDs: []string{"getCatStatus"},
				},
			},
		}

		// Run our code generation:
		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.NotContains(t, code.GetCombined(), `"/cat"`)
	})

	t.Run("examples removed", func(t *testing.T) {
		contents, err := os.ReadFile("testdata/with-examples.yml")
		require.NoError(t, err)

		opts := Configuration{
			PackageName: packageName,
			Output: &Output{
				UseSingleFile: true,
			},
		}

		// Run our code generation:
		code, err := Generate(contents, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		combined := code.GetCombined()
		assert.Contains(t, combined, `type ExtraA = string`)
		assert.Contains(t, combined, `type ExtraB = string`)

		// Load the document to verify examples are removed
		doc, err := LoadDocumentFromContents(contents)
		require.NoError(t, err)

		// Apply filtering (which includes example removal)
		filteredDoc, err := filterOutDocument(doc, opts.Filter)
		require.NoError(t, err)

		model, err := filteredDoc.BuildV3Model()
		require.NoError(t, err)

		// components/examples should be empty (no example references)
		assert.Equal(t, 0, model.Model.Components.Examples.Len())

		// Schema examples should be empty, but example (singular) should be kept
		paySessionReq := model.Model.Components.Schemas.GetOrZero("PaySessionRequest")
		schema := paySessionReq.Schema()
		assert.Nil(t, schema.Examples)
		assert.NotNil(t, schema.Example)

		// Property examples should be empty, but example (singular) should be kept
		fooProp := schema.Properties.GetOrZero("foo")
		fooSchema := fooProp.Schema()
		assert.Nil(t, fooSchema.Examples)
		assert.NotNil(t, fooSchema.Example)

		// Request body examples should be empty
		reqBody := model.Model.Components.RequestBodies.GetOrZero("PaySessionRequest")
		jsonContent := reqBody.Content.GetOrZero("application/json")
		assert.Equal(t, 0, jsonContent.Examples.Len())

		// XML content examples should be empty, but example (singular) should be kept
		xmlContent := reqBody.Content.GetOrZero("application/xml")
		assert.Equal(t, 0, xmlContent.Examples.Len())
		assert.NotNil(t, xmlContent.Example)

		// Parameter examples should be empty, but example (singular) should be kept
		param := model.Model.Components.Parameters.GetOrZero("Idempotency-Key")
		assert.Equal(t, 0, param.Examples.Len())
		assert.NotNil(t, param.Example)

		// Header examples should be empty, but example (singular) should be kept
		header := model.Model.Components.Headers.GetOrZero("Idempotency-Key")
		assert.Equal(t, 0, header.Examples.Len())
		assert.NotNil(t, header.Example)

		// Response examples should be empty
		response := model.Model.Components.Responses.GetOrZero("SuccessResponse")
		jsonRespContent := response.Content.GetOrZero("application/json")
		assert.Equal(t, 0, jsonRespContent.Examples.Len())
	})

	t.Run("examples removed from webhooks", func(t *testing.T) {
		contents, err := os.ReadFile("testdata/webhooks-with-examples.yml")
		require.NoError(t, err)

		opts := Configuration{
			PackageName: packageName,
			Output: &Output{
				UseSingleFile: true,
			},
		}

		// Run our code generation - should not error
		code, err := Generate(contents, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		// Load the document to verify examples are removed
		doc, err := LoadDocumentFromContents(contents)
		require.NoError(t, err)

		// Apply filtering (which includes example removal)
		filteredDoc, err := filterOutDocument(doc, opts.Filter)
		require.NoError(t, err)

		model, err := filteredDoc.BuildV3Model()
		require.NoError(t, err)

		// components/examples should be empty
		assert.Equal(t, 0, model.Model.Components.Examples.Len())
	})

	t.Run("empty filter does not remove operations", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Include: FilterParamsConfig{
					OperationIDs: []string{},
				},
				Exclude: FilterParamsConfig{
					OperationIDs: []string{},
				},
			},
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		// Run our code generation:
		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		// Should contain operations since empty filters don't remove anything
		combined := code.GetCombined()
		assert.Contains(t, combined, `type CatDeadCause string`)
		assert.Contains(t, combined, `"/cat"`)
	})

	t.Run("nil filter does not remove operations", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Include: FilterParamsConfig{
					OperationIDs: nil,
				},
				Exclude: FilterParamsConfig{
					OperationIDs: nil,
				},
			},
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		// Run our code generation:
		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		// Should contain operations since nil filters don't remove anything
		combined := code.GetCombined()
		assert.Contains(t, combined, `type CatDeadCause string`)
		assert.Contains(t, combined, `"/cat"`)
	})

	t.Run("default configuration does not filter operations", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			// Filter is zero value - should not filter anything
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		// Run our code generation:
		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		// Should contain all operations
		combined := code.GetCombined()
		assert.Contains(t, combined, `type CatDeadCause string`)
		assert.Contains(t, combined, `"/cat"`)
	})
}

func TestFilterOperationsByPath(t *testing.T) {
	packageName := "testswagger"

	t.Run("include paths", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Include: FilterParamsConfig{
					Paths: []string{"/cat"},
				},
			},
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		combined := code.GetCombined()
		assert.Contains(t, combined, `"/cat"`)
		assert.NotContains(t, combined, `"/test/{name}"`)
		assert.NotContains(t, combined, `"/enum"`)
	})

	t.Run("exclude paths", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Exclude: FilterParamsConfig{
					Paths: []string{"/cat", "/enum"},
				},
			},
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		combined := code.GetCombined()
		assert.NotContains(t, combined, `"/cat"`)
		assert.NotContains(t, combined, `"/enum"`)
		assert.Contains(t, combined, `"/test/{name}"`)
	})

	t.Run("empty include paths does not filter", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Include: FilterParamsConfig{
					Paths: []string{}, // Empty - should not filter
				},
			},
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		combined := code.GetCombined()
		assert.Contains(t, combined, `"/cat"`)
		assert.Contains(t, combined, `"/test/{name}"`)
		assert.Contains(t, combined, `"/enum"`)
	})

	t.Run("empty exclude paths does not filter", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Exclude: FilterParamsConfig{
					Paths: []string{}, // Empty - should not filter
				},
			},
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		combined := code.GetCombined()
		assert.Contains(t, combined, `"/cat"`)
		assert.Contains(t, combined, `"/test/{name}"`)
		assert.Contains(t, combined, `"/enum"`)
	})

	t.Run("nil include paths does not filter", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Include: FilterParamsConfig{
					Paths: nil, // Nil - should not filter
				},
			},
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		combined := code.GetCombined()
		assert.Contains(t, combined, `"/cat"`)
		assert.Contains(t, combined, `"/test/{name}"`)
		assert.Contains(t, combined, `"/enum"`)
	})

	t.Run("nil exclude paths does not filter", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Filter: FilterConfig{
				Exclude: FilterParamsConfig{
					Paths: nil, // Nil - should not filter
				},
			},
			Generate: &GenerateOptions{
				Client: true,
			},
			Output: &Output{
				UseSingleFile: true,
			},
		}

		code, err := Generate([]byte(testDocument), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, code)

		combined := code.GetCombined()
		assert.Contains(t, combined, `"/cat"`)
		assert.Contains(t, combined, `"/test/{name}"`)
		assert.Contains(t, combined, `"/enum"`)
	})
}
