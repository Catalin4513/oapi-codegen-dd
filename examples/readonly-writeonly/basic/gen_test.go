package gen

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateUserBody_ReadOnlyFieldsOptional(t *testing.T) {
	// Test that we can create a request body without readOnly fields
	body := CreateUserBody{
		Name:  "John Doe",
		Email: "john@example.com",
		// ID and CreatedAt are readOnly and should be optional
	}

	// Validate should pass even without readOnly fields
	err := body.Validate()
	require.NoError(t, err)

	// Marshal to JSON
	data, err := json.Marshal(body)
	require.NoError(t, err)

	// Should only contain name and email, not id or createdAt
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	require.Equal(t, "John Doe", result["name"])
	require.Equal(t, "john@example.com", result["email"])
	require.NotContains(t, result, "id")
	require.NotContains(t, result, "createdAt")
}

func TestCreateUserResponse_ReadOnlyFieldsRequired(t *testing.T) {
	// Response should have all fields including readOnly ones
	jsonData := `{
		"id": "123",
		"name": "John Doe",
		"email": "john@example.com",
		"createdAt": "2024-01-01T00:00:00Z"
	}`

	var response CreateUserResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	require.NoError(t, err)

	require.Equal(t, "123", response.ID)
	require.Equal(t, "John Doe", response.Name)
	require.Equal(t, "john@example.com", response.Email)
}

func TestUser_ComponentSchema(t *testing.T) {
	// The component schema should still have all fields required
	jsonData := `{
		"id": "123",
		"name": "John Doe",
		"email": "john@example.com",
		"createdAt": "2024-01-01T00:00:00Z"
	}`

	var user User
	err := json.Unmarshal([]byte(jsonData), &user)
	require.NoError(t, err)

	err = user.Validate()
	require.NoError(t, err)

	require.Equal(t, "123", user.ID)
	require.Equal(t, "John Doe", user.Name)
	require.Equal(t, "john@example.com", user.Email)
}
