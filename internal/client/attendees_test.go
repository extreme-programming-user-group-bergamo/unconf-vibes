package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAttendees_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/conferences/socrates-26/attendees", r.URL.Path)
		assert.Equal(t, "Bearer valid-access-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AttendeeListResponse{
			Attendees: []AttendeeProjectionResponse{
				{
					DisplayName:    "Alice",
					GitHubUsername: "alice",
					Room: &AttendeeRoomProjectionResponse{
						RoomNumber: "101",
						RoomType:   "double",
					},
				},
			},
			PrivateAttendeesCount: 2,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ListAttendees(context.Background(), "valid-access-token", "socrates-26")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Attendees, 1)
	assert.Equal(t, "Alice", resp.Attendees[0].DisplayName)
	assert.Equal(t, 2, resp.PrivateAttendeesCount)
}

func TestListAttendees_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "unauthorized",
				"message": "Unauthorized",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ListAttendees(context.Background(), "invalid-token", "socrates-26")
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestListAttendees_ConferenceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "not_found",
				"message": "Conference not found",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ListAttendees(context.Background(), "valid-token", "missing")
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConferenceNotFound)
}

func TestListAttendees_NilAttendeesDefaultsToEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"private_attendees_count": 0,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ListAttendees(context.Background(), "valid-token", "socrates-26")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Attendees)
	assert.Empty(t, resp.Attendees)
}
