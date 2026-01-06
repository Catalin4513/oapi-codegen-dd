// Copyright 2025 DoorDash, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestRegisterCustomTypeFunc_WithEither(t *testing.T) {
	v := validator.New(validator.WithRequiredStructEnabled())
	RegisterCustomTypeFunc(v)

	type TestStruct struct {
		Value Either[string, int] `validate:"required"`
	}

	t.Run("validates active A variant", func(t *testing.T) {
		ts := TestStruct{
			Value: NewEitherFromA[string, int]("hello"),
		}
		err := v.Struct(ts)
		assert.NoError(t, err)
	})

	t.Run("validates active B variant", func(t *testing.T) {
		ts := TestStruct{
			Value: NewEitherFromB[string, int](42),
		}
		err := v.Struct(ts)
		assert.NoError(t, err)
	})

	t.Run("fails validation when no variant is active", func(t *testing.T) {
		ts := TestStruct{
			Value: Either[string, int]{}, // N=0, no active variant
		}
		err := v.Struct(ts)
		assert.Error(t, err)
	})
}

func TestRegisterCustomTypeFunc_WithValidateVar(t *testing.T) {
	v := validator.New(validator.WithRequiredStructEnabled())
	RegisterCustomTypeFunc(v)

	t.Run("validates active A variant with Var", func(t *testing.T) {
		either := NewEitherFromA[string, int]("hello")
		err := v.Var(either, "required")
		assert.NoError(t, err)
	})

	t.Run("validates active B variant with Var", func(t *testing.T) {
		either := NewEitherFromB[string, int](42)
		err := v.Var(either, "required")
		assert.NoError(t, err)
	})

	t.Run("returns nil when no variant is active with Var", func(t *testing.T) {
		either := Either[string, int]{} // N=0, no active variant
		// The custom type function returns nil for inactive variants,
		// which the validator treats as a zero value, not a validation error
		err := v.Var(either, "required")
		// This doesn't error because the validator sees nil from Value()
		// Actual validation of inactive variants should be done via Validate() method
		assert.NoError(t, err)
	})
}
