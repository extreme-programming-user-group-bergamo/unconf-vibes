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

func TestListRooms_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/conferences/socrates-26/rooms", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]RoomResponse{
			{RoomNumber: "101", RoomType: "double", Capacity: 2, SpotsAvailable: 1},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rooms, err := c.ListRooms(context.Background(), "socrates-26")

	require.NoError(t, err)
	require.Len(t, rooms, 1)
	assert.Equal(t, "101", rooms[0].RoomNumber)
}

func TestListRooms_ConferenceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "not_found",
				"message": "conference not found",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rooms, err := c.ListRooms(context.Background(), "missing")

	assert.Nil(t, rooms)
	assert.ErrorIs(t, err, ErrConferenceNotFound)
}

func TestListRooms_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "internal_error",
				"message": "something broke",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rooms, err := c.ListRooms(context.Background(), "socrates-26")

	assert.Nil(t, rooms)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list rooms")
}

func TestCreateRoom_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "conflict",
				"message": "room exists",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	room, err := c.CreateRoom(context.Background(), "token", "socrates-26", ManageRoomRequest{
		RoomNumber:    "101",
		RoomType:      "double",
		PricePerNight: 100,
		Capacity:      2,
	})
	assert.Nil(t, room)
	assert.ErrorIs(t, err, ErrRoomExists)
}

func TestDeleteRoom_HasBookings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "conflict",
				"message": "room has bookings",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.DeleteRoom(context.Background(), "token", "socrates-26", "101")
	assert.ErrorIs(t, err, ErrRoomHasBookings)
}
