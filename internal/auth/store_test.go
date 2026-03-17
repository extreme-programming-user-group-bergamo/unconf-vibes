package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockTokenStore_SaveAndRetrieve(t *testing.T) {
	store := NewMockTokenStore()

	err := store.SaveTokens("access123", "refresh456")
	require.NoError(t, err)

	access, err := store.GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "access123", access)

	refresh, err := store.GetRefreshToken()
	require.NoError(t, err)
	assert.Equal(t, "refresh456", refresh)
}

func TestMockTokenStore_HasValidToken(t *testing.T) {
	store := NewMockTokenStore()

	assert.False(t, store.HasValidToken())

	err := store.SaveTokens("token", "refresh")
	require.NoError(t, err)

	assert.True(t, store.HasValidToken())
}

func TestMockTokenStore_ClearTokens(t *testing.T) {
	store := NewMockTokenStore()

	err := store.SaveTokens("token", "refresh")
	require.NoError(t, err)
	assert.True(t, store.HasValidToken())

	err = store.ClearTokens()
	require.NoError(t, err)
	assert.False(t, store.HasValidToken())

	_, err = store.GetAccessToken()
	assert.ErrorIs(t, err, ErrNotAuthenticated)
}

func TestMockTokenStore_ErrorPropagation(t *testing.T) {
	testErr := assert.AnError
	store := NewMockTokenStoreWithError(testErr)

	err := store.SaveTokens("a", "b")
	assert.ErrorIs(t, err, testErr)

	_, err = store.GetAccessToken()
	assert.ErrorIs(t, err, testErr)

	_, err = store.GetRefreshToken()
	assert.ErrorIs(t, err, testErr)

	err = store.ClearTokens()
	assert.ErrorIs(t, err, testErr)
}

func TestRequireAuth_NotAuthenticated(t *testing.T) {
	store := NewMockTokenStore()
	err := RequireAuth(store)
	assert.ErrorIs(t, err, ErrNotAuthenticated)
}

func TestRequireAuth_Authenticated(t *testing.T) {
	store := NewMockTokenStore()
	store.SetTokens("valid-token", "refresh-token")

	err := RequireAuth(store)
	assert.NoError(t, err)
}
