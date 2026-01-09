// Copyright 2025 DoorDash, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package runtime

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsMap(t *testing.T) {
	t.Run("converts struct to map[string]any", func(t *testing.T) {
		type TestStruct struct {
			Name  string `json:"name"`
			Age   int    `json:"age"`
			Email string `json:"email,omitempty"`
		}

		input := TestStruct{Name: "John", Age: 30}
		result, err := AsMap[any](input)

		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, float64(30), result["age"]) // JSON numbers unmarshal as float64
		assert.NotContains(t, result, "email")      // omitempty field not present
	})

	t.Run("converts struct to map[string]string", func(t *testing.T) {
		type HeaderStruct struct {
			Authorization string `json:"authorization"`
			ContentType   string `json:"content-type"`
		}

		input := HeaderStruct{Authorization: "Bearer token", ContentType: "application/json"}
		result, err := AsMap[string](input)

		require.NoError(t, err)
		assert.Equal(t, "Bearer token", result["authorization"])
		assert.Equal(t, "application/json", result["content-type"])
	})

	t.Run("returns nil for nil input", func(t *testing.T) {
		result, err := AsMap[any](nil)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("handles pointer fields", func(t *testing.T) {
		type TestStruct struct {
			Name  *string `json:"name,omitempty"`
			Count *int    `json:"count,omitempty"`
		}

		name := "test"
		count := 42
		input := TestStruct{Name: &name, Count: &count}
		result, err := AsMap[any](input)

		require.NoError(t, err)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, float64(42), result["count"])
	})

	t.Run("handles nested structs", func(t *testing.T) {
		type Address struct {
			City string `json:"city"`
		}
		type Person struct {
			Name    string  `json:"name"`
			Address Address `json:"address"`
		}

		input := Person{Name: "John", Address: Address{City: "NYC"}}
		result, err := AsMap[any](input)

		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
		address, ok := result["address"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "NYC", address["city"])
	})
}

var (
	outputJSON, outputIndentJSON, outputNonexistentJSON string
	input                                               = `
{  
  "number": 1,
  "string": "value",
  "object": {
    "number": 1,
    "string": "value",
    "nested_object": {
      "number": 2
    },
    "array": [1, 2, 3],
    "partial_array": [1, 2, 3]
  }
}
    `
	patch = `
{  
  "number": 2,
  "string": "value1",
  "nonexitent": "woot",
  "object": {
    "number": 3,
    "string": "value2",
    "nested_object": {
      "number": 4
    },
    "array": [3, 2, 1],
    "partial_array": {
      "1": 4
    }
  }
}
    `
)

func TestMain(t *testing.M) {
	output := []byte(`
{  
  "number": 2,
  "string": "value1",
  "object": {
    "number": 3,
    "string": "value2",
    "nested_object": {
      "number": 4
    },
    "array": [3, 2, 1],
    "partial_array": [1, 4, 3]
  }
}
    `)
	outputNonexistent := []byte(`
{
  "number": 2,
  "string": "value1",
  "nonexitent": "woot",
  "object": {
    "number": 3,
    "string": "value2",
    "nested_object": {
      "number": 4
    },
    "array": [3, 2, 1],
    "partial_array": [1, 4, 3]
  }
}
`)

	var outputData any
	_ = json.Unmarshal(output, &outputData)

	output, _ = json.Marshal(outputData)
	outputJSON = string(output)

	output, _ = json.MarshalIndent(outputData, " ", "  ")
	outputIndentJSON = string(output)

	var outputNonexistentData any
	_ = json.Unmarshal(outputNonexistent, &outputNonexistentData)
	output, _ = json.Marshal(outputNonexistentData)
	outputNonexistentJSON = string(output)

	t.Run()
}

func TestMergeBytesIndent(t *testing.T) {
	merger := &Merger{}
	result, err := merger.MergeBytesIndent([]byte(input), []byte(patch), " ", "  ")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if string(result) != outputIndentJSON {
		t.Errorf("Result not equals output\nExpected:\n%s\n\nGot:\n%s\n\n", outputIndentJSON, result)
	}

	if len(merger.Errors) != 0 {
		t.Errorf("info.Errors count is not 0, count: %v", len(merger.Errors))
	}
}

func TestMergeBytes(t *testing.T) {
	merger := &Merger{}
	result, err := merger.MergeBytes([]byte(input), []byte(patch))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if string(result) != outputJSON {
		t.Errorf("Result not equals output\nExpected:\n%s\n\nGot:\n%s\n\n", outputJSON, result)
	}

	if len(merger.Errors) != 0 {
		t.Errorf("info.Errors count is not 0, count: %v", len(merger.Errors))
	}
}

func TestMergeBytesNonexistent(t *testing.T) {
	merger := &Merger{
		CopyNonexistent: true,
	}
	result, err := merger.MergeBytes([]byte(input), []byte(patch))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if string(result) != outputNonexistentJSON {
		t.Errorf("Result not equals output\nExpected:\n%s\n\nGot:\n%s\n\n", outputNonexistentJSON, result)
	}

	if len(merger.Errors) != 0 {
		t.Errorf("info.Errors count is not 0, count: %v", len(merger.Errors))
	}
}

func TestLongNumbers(t *testing.T) {
	input := `{"Id":12423434,"Value":12423434}`
	patch := `{"Value":12423439}`
	outputJSON := `{"Id":12423434,"Value":12423439}`

	merger := &Merger{}
	result, err := merger.MergeBytes([]byte(input), []byte(patch))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if string(result) != outputJSON {
		t.Errorf("Result not equals output\nExpected:\n%s\n\nGot:\n%s\n\n", outputJSON, result)
	}

	if len(merger.Errors) != 0 {
		t.Errorf("info.Errors count is not 0, count: %v", len(merger.Errors))
	}
}

func TestJSONMerge(t *testing.T) {
	t.Run("when object", func(t *testing.T) {
		t.Run("Merges properties defined in both objects", func(t *testing.T) {
			data := `{"foo": 1}`
			patch := `{"foo": null}`
			expected := `{"foo":null}`

			actual, err := JSONMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Sets property defined in only src object", func(t *testing.T) {
			data := `{}`
			patch := `{"source":"merge-me"}`
			expected := `{"source":"merge-me"}`

			actual, err := JSONMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Handles child objects", func(t *testing.T) {
			data := `{"channel":{"status":"valid"}}`
			patch := `{"channel":{"id":1}}`
			expected := `{"channel":{"id":1,"status":"valid"}}`

			actual, err := JSONMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Handles empty objects", func(t *testing.T) {
			data := `{}`
			patch := `{}`
			expected := `{}`

			actual, err := JSONMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Handles nil data", func(t *testing.T) {
			patch := `{"foo":"bar"}`
			expected := `{"foo":"bar"}`

			actual, err := JSONMerge(nil, []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Handles nil patch", func(t *testing.T) {
			data := `{"foo":"bar"}`
			expected := `{"foo":"bar"}`

			actual, err := JSONMerge([]byte(data), nil)
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})
	})
	t.Run("when array", func(t *testing.T) {
		t.Run("it does not merge", func(t *testing.T) {
			data := `[{"foo": 1}]`
			patch := `[{"foo": null}]`
			expected := `[{"foo":1}]`

			actual, err := JSONMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})
	})
}

func TestCoalesceOrMerge(t *testing.T) {
	t.Run("when object", func(t *testing.T) {
		parts := []json.RawMessage{
			[]byte(`{"foo": 1}`),
			[]byte(`{"bar": {"car": "var"}}`),
			[]byte(`{"baz": null}`),
		}

		expected := `{"bar":{"car":"var"},"baz":null,"foo":1}`

		res, err := CoalesceOrMerge(parts...)
		require.NoError(t, err)
		require.JSONEq(t, expected, string(res))
	})

	t.Run("when array", func(t *testing.T) {
		parts := []json.RawMessage{
			[]byte(`[1, 2]`),
			[]byte(`[3, 4]`),
		}
		expected := `[1,2,3,4]`
		res, err := CoalesceOrMerge(parts...)
		require.NoError(t, err)
		require.JSONEq(t, expected, string(res))
	})

	t.Run("when scalar", func(t *testing.T) {
		parts := []json.RawMessage{
			[]byte(`1`),
			[]byte(`2`),
		}

		res, err := CoalesceOrMerge(parts...)
		assert.Nil(t, res)
		assert.Equal(t, "cannot combine 2 non-null branches of mixed/unsupported kinds", err.Error())
	})
}
