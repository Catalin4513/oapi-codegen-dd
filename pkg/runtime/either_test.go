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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEitherFromA(t *testing.T) {
	res := NewEitherFromA[string, int]("test")

	assert.True(t, res.IsA())
	assert.Equal(t, "test", res.A)
	assert.False(t, res.IsB())
	assert.Equal(t, 1, res.N)
	assert.Equal(t, 0, res.B)
}

func TestNewEitherFromB(t *testing.T) {
	res := NewEitherFromB[string, int](10)

	assert.False(t, res.IsA())
	assert.Equal(t, "", res.A)
	assert.True(t, res.IsB())
	assert.Equal(t, 2, res.N)
	assert.Equal(t, 10, res.B)
}

func TestEither_Value(t *testing.T) {
	res := NewEitherFromA[string, int]("test")
	assert.Equal(t, "test", res.Value())

	res = NewEitherFromB[string, int](10)
	assert.Equal(t, 10, res.Value())
}

func TestEither_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Either[string, int]
	}{
		{
			name:     "string",
			input:    []byte(`"test"`),
			expected: NewEitherFromA[string, int]("test"),
		},
		{
			name:     "int",
			input:    []byte(`10`),
			expected: NewEitherFromB[string, int](10),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res Either[string, int]
			err := res.UnmarshalJSON(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, res)
		})
	}
}

type NameOrID struct {
	Either[IDWrapper, NameWrapper]
}

type IDWrapper struct {
	ID int `json:"id"`
}

type NameWrapper struct {
	Name string `json:"name"`
}

func TestEither_MarshalJSON_with_wrapper(t *testing.T) {
	tests := []struct {
		name     string
		input    NameOrID
		expected []byte
	}{
		{
			name:     "id",
			input:    NameOrID{Either: NewEitherFromA[IDWrapper, NameWrapper](IDWrapper{ID: 10})},
			expected: []byte(`{"id":10}`),
		},
		{
			name:     "name",
			input:    NameOrID{Either: NewEitherFromB[IDWrapper, NameWrapper](NameWrapper{Name: "test"})},
			expected: []byte(`{"name":"test"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := test.input.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, string(test.expected), string(res))
		})
	}
}

func TestEither_UnmarshalJSON_with_wrapper(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected NameOrID
	}{
		{
			name:     "id",
			input:    []byte(`{"id":10}`),
			expected: NameOrID{Either: NewEitherFromA[IDWrapper, NameWrapper](IDWrapper{ID: 10})},
		},
		{
			name:     "name",
			input:    []byte(`{"name":"test"}`),
			expected: NameOrID{Either: NewEitherFromB[IDWrapper, NameWrapper](NameWrapper{Name: "test"})},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res NameOrID
			err := res.UnmarshalJSON(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, res)
		})
	}
}

type ValidatableStruct struct {
	Name string `validate:"required"`
	Age  int    `validate:"required,min=1"`
}

func (v ValidatableStruct) Validate() error {
	if v.Name == "" {
		return assert.AnError
	}
	if v.Age < 1 {
		return assert.AnError
	}
	return nil
}

func TestEither_Validate(t *testing.T) {
	t.Run("validates A variant when active", func(t *testing.T) {
		valid := ValidatableStruct{Name: "John", Age: 30}
		either := NewEitherFromA[ValidatableStruct, string](valid)

		err := either.Validate()
		assert.NoError(t, err)
	})

	t.Run("validates B variant when active", func(t *testing.T) {
		either := NewEitherFromB[ValidatableStruct, string]("test")

		err := either.Validate()
		assert.NoError(t, err)
	})

	t.Run("fails validation for invalid A variant", func(t *testing.T) {
		invalid := ValidatableStruct{Name: "", Age: 0}
		either := NewEitherFromA[ValidatableStruct, string](invalid)

		err := either.Validate()
		assert.Error(t, err)
	})

	t.Run("does not validate inactive B variant", func(t *testing.T) {
		valid := ValidatableStruct{Name: "John", Age: 30}
		either := NewEitherFromA[ValidatableStruct, string](valid)
		// B is inactive and would be invalid if checked (empty string)

		err := either.Validate()
		assert.NoError(t, err) // Should pass because only A is validated
	})
}
