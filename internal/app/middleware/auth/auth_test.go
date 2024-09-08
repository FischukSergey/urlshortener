package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestBuildJWTString(t *testing.T) {
	token, userID, err := BuildJWTString()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, userID, 0)

	// Verify the token
	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
	assert.Equal(t, userID, claims.UserID)
	assert.WithinDuration(t, time.Now().Add(TokenExp), claims.ExpiresAt.Time, time.Second)
}

func TestGetUserID(t *testing.T) {
	// Test valid token
	token, userID, _ := BuildJWTString()
	retrievedID := GetUserID(token)
	assert.Equal(t, userID, retrievedID)

	// Test invalid token
	invalidID := GetUserID("invalid.token.string")
	assert.Equal(t, -1, invalidID)
}
