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

func TestListRoommateRequests_SuccessIncludesProjectionFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/requests", r.URL.Path)
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]RoommateRequestResponse{
			{
				ID:            7,
				Direction:     "incoming",
				RequesterName: "Alice",
				RoomNumber:    "101",
				RoomType:      "double",
				Status:        "pending",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	result, err := c.ListRoommateRequests(context.Background(), "token")

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Alice", result[0].RequesterName)
	assert.Equal(t, "101", result[0].RoomNumber)
	assert.Equal(t, "double", result[0].RoomType)
}

func TestRespondToRoommateRequest_ErrorMappings(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
		action     string
		wantErr    error
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, action: "accept", wantErr: ErrUnauthorized},
		{name: "not found", statusCode: http.StatusNotFound, errorCode: "not_found", action: "accept", wantErr: ErrRequestNotFound},
		{name: "forbidden", statusCode: http.StatusForbidden, errorCode: "forbidden", action: "accept", wantErr: ErrRequestForbidden},
		{name: "invalid state", statusCode: http.StatusConflict, errorCode: "invalid_request_state", action: "decline", wantErr: ErrInvalidRequestState},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    tc.errorCode,
						"message": "boom",
					},
				})
			}))
			defer srv.Close()

			c := NewClient(srv.URL)
			var err error
			if tc.action == "accept" {
				_, err = c.AcceptRoommateRequest(context.Background(), "token", 42)
			} else {
				_, err = c.DeclineRoommateRequest(context.Background(), "token", 42)
			}

			require.Error(t, err)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
