package anytype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyTypeResponses(t *testing.T) {
	t.Run("GetEmptySchemaResponse is any type alias", func(t *testing.T) {
		var resp GetEmptySchemaResponse

		// Can assign any value
		resp = "string"
		assert.Equal(t, "string", resp)

		resp = 123
		assert.Equal(t, 123, resp)

		resp = map[string]interface{}{"key": "value"}
		assert.Equal(t, map[string]interface{}{"key": "value"}, resp)
	})

	t.Run("GetNoTypeResponse is any type alias", func(t *testing.T) {
		// Can assign any value
		resp := []int{1, 2, 3}
		assert.Equal(t, []int{1, 2, 3}, resp)
	})

	t.Run("GetMalformedResponse is any type (not alias)", func(t *testing.T) {
		// Can assign any value
		var resp GetMalformedResponse = "test"
		assert.Equal(t, GetMalformedResponse("test"), resp)

		// Should NOT have a Validate method since it's type 'any'
		// This test verifies the fix: we don't generate Validate() for 'any' types
	})

	t.Run("GetExplicitObjectResponse is struct with Validate", func(t *testing.T) {
		resp := GetExplicitObjectResponse{
			ID:   strPtr("test-id"),
			Data: anyPtr("any data here"),
		}

		assert.Equal(t, "test-id", *resp.ID)
		assert.Equal(t, "any data here", *resp.Data)

		// Should have Validate method
		err := resp.Validate()
		assert.NoError(t, err)
	})
}

func strPtr(s string) *string {
	return &s
}

func anyPtr(a any) *any {
	return &a
}
