package gen

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeasurement_Value_StringOrNumber(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		data := []byte(`{"value": "hello"}`)
		var m Measurement
		err := json.Unmarshal(data, &m)

		require.NoError(t, err)
		require.NotNil(t, m.Value)
		assert.True(t, m.Value.IsA())
		assert.Equal(t, "hello", m.Value.A)
	})

	t.Run("number value", func(t *testing.T) {
		data := []byte(`{"value": 42.5}`)
		var m Measurement
		err := json.Unmarshal(data, &m)

		require.NoError(t, err)
		require.NotNil(t, m.Value)
		assert.True(t, m.Value.IsB())
		assert.Equal(t, float32(42.5), m.Value.B)
	})
}

func TestFlexibleID_StringOrInteger(t *testing.T) {
	t.Run("string id", func(t *testing.T) {
		data := []byte(`"abc-123"`)
		var id FlexibleID
		err := json.Unmarshal(data, &id)

		require.NoError(t, err)
		assert.True(t, id.IsA())
		assert.Equal(t, "abc-123", id.A)
	})

	t.Run("integer id", func(t *testing.T) {
		data := []byte(`12345`)
		var id FlexibleID
		err := json.Unmarshal(data, &id)

		require.NoError(t, err)
		assert.True(t, id.IsB())
		assert.Equal(t, 12345, id.B)
	})
}
