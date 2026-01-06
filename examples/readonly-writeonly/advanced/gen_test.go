package advanced

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUserBody_ReadOnlyFieldsOptional(t *testing.T) {
	// Test that readOnly fields (id, createdAt) are optional in request bodies
	// even though they are marked as required in the schema

	requestJSON := `{
		"name": "John Doe",
		"email": "john@example.com",
		"password": "secret123",
		"bio": "Software engineer"
	}`

	var body CreateUserBody
	err := json.Unmarshal([]byte(requestJSON), &body)
	require.NoError(t, err)

	// Verify that readOnly fields are nil (not required)
	assert.Nil(t, body.ID, "ID should be optional in request body")
	assert.Nil(t, body.CreatedAt, "CreatedAt should be optional in request body")

	// Verify that regular required fields are present
	assert.Equal(t, "John Doe", body.Name)
	assert.Equal(t, "john@example.com", string(body.Email))
	assert.Equal(t, "secret123", body.Password)
	assert.NotNil(t, body.Bio)
	assert.Equal(t, "Software engineer", *body.Bio)
}

func TestUpdateUserBody_ReadOnlyFieldsOptional(t *testing.T) {
	// Test PATCH request body - readOnly fields should be optional

	requestJSON := `{
		"name": "Jane Doe",
		"email": "jane@example.com"
	}`

	var body UpdateUserBody
	err := json.Unmarshal([]byte(requestJSON), &body)
	require.NoError(t, err)

	// Verify that readOnly fields are nil
	assert.Nil(t, body.ID, "ID should be optional in PATCH request")
	assert.Nil(t, body.CreatedAt, "CreatedAt should be optional in PATCH request")

	// Verify that we can update just some fields
	assert.Equal(t, "Jane Doe", body.Name)
	assert.Equal(t, "jane@example.com", string(body.Email))
}

func TestUser_ReadOnlyFieldsRequired(t *testing.T) {
	// Test that readOnly fields (id, createdAt) are required in responses

	now := time.Now()
	responseJSON := `{
		"id": "user-123",
		"name": "John Doe",
		"email": "john@example.com",
		"password": "secret123",
		"createdAt": "` + now.Format(time.RFC3339) + `"
	}`

	var user User
	err := json.Unmarshal([]byte(responseJSON), &user)
	require.NoError(t, err)

	// Verify that readOnly fields are non-pointer (required)
	assert.Equal(t, "user-123", user.ID)
	assert.False(t, user.CreatedAt.IsZero())

	// Verify regular fields
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", string(user.Email))
}

func TestUser_OptionalReadOnlyFields(t *testing.T) {
	// Test that readOnly fields that are NOT required remain optional

	responseJSON := `{
		"id": "user-123",
		"name": "John Doe",
		"email": "john@example.com",
		"password": "secret123",
		"createdAt": "2024-01-01T00:00:00Z"
	}`

	var user User
	err := json.Unmarshal([]byte(responseJSON), &user)
	require.NoError(t, err)

	// updatedAt and lastLogin are readOnly but NOT required, so they should be nil
	assert.Nil(t, user.UpdatedAt, "UpdatedAt should be optional even in responses")
	assert.Nil(t, user.LastLogin, "LastLogin should be optional even in responses")
}

func TestWriteOnlyFieldRequired(t *testing.T) {
	// Test that writeOnly field (password) is required in request bodies

	requestJSON := `{
		"name": "John Doe",
		"email": "john@example.com"
	}`

	var body CreateUserBody
	err := json.Unmarshal([]byte(requestJSON), &body)
	require.NoError(t, err)

	// Password is writeOnly + required, so it should be empty string (not provided)
	assert.Equal(t, "", body.Password, "Password was not provided in JSON")

	// With password
	requestWithPassword := `{
		"name": "John Doe",
		"email": "john@example.com",
		"password": "secret123"
	}`

	err = json.Unmarshal([]byte(requestWithPassword), &body)
	require.NoError(t, err)
	assert.Equal(t, "secret123", body.Password)
}
