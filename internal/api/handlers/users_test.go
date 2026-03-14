package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockUserService struct {
	getProfileFn    func(ctx context.Context, userID int64) (*models.User, error)
	updateProfileFn func(ctx context.Context, userID int64, input service.UpdateProfileInput) (*models.User, error)
}

func (m *mockUserService) GetProfile(ctx context.Context, userID int64) (*models.User, error) {
	if m.getProfileFn != nil {
		return m.getProfileFn(ctx, userID)
	}

	return nil, nil
}

func (m *mockUserService) UpdateProfile(ctx context.Context, userID int64, input service.UpdateProfileInput) (*models.User, error) {
	if m.updateProfileFn != nil {
		return m.updateProfileFn(ctx, userID, input)
	}

	return nil, nil
}

func setupUserRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/users/me", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.GetMe(c)
	})
	router.PUT("/users/me", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.UpdateMe(c)
	})

	return router
}

func testUser() *models.User {
	return &models.User{
		ID:             42,
		GitHubID:       "12345",
		Email:          "test@example.com",
		DisplayName:    "Test User",
		PrivacySetting: "public",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
}

func TestUserHandler_GetMe_Success(t *testing.T) {
	user := testUser()
	handler := NewUserHandler(&mockUserService{
		getProfileFn: func(_ context.Context, userID int64) (*models.User, error) {
			assert.Equal(t, int64(42), userID)
			return user, nil
		},
	})

	router := setupUserRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"display_name":"Test User"`)
	assert.Contains(t, w.Body.String(), `"email":"test@example.com"`)
}

func TestUserHandler_GetMe_NotFound(t *testing.T) {
	handler := NewUserHandler(&mockUserService{
		getProfileFn: func(_ context.Context, _ int64) (*models.User, error) {
			return nil, service.ErrUserNotFound
		},
	})

	router := setupUserRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user_not_found")
}

func TestUserHandler_UpdateMe_Success(t *testing.T) {
	user := testUser()
	user.DisplayName = "Updated Name"

	handler := NewUserHandler(&mockUserService{
		updateProfileFn: func(_ context.Context, userID int64, input service.UpdateProfileInput) (*models.User, error) {
			assert.Equal(t, int64(42), userID)
			assert.Equal(t, "Updated Name", *input.DisplayName)
			return user, nil
		},
	})

	router := setupUserRouter(handler)
	body := `{"display_name":"Updated Name"}`
	req := httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"display_name":"Updated Name"`)
}

func TestUserHandler_UpdateMe_InvalidPrivacySetting(t *testing.T) {
	handler := NewUserHandler(&mockUserService{
		updateProfileFn: func(_ context.Context, _ int64, _ service.UpdateProfileInput) (*models.User, error) {
			return nil, service.ErrInvalidPrivacySetting
		},
	})

	router := setupUserRouter(handler)
	body := `{"privacy_setting":"invalid"}`
	req := httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_privacy_setting")
}

func TestUserHandler_UpdateMe_InvalidBody(t *testing.T) {
	handler := NewUserHandler(&mockUserService{})

	router := setupUserRouter(handler)
	req := httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}

func TestUserHandler_UpdateMe_UserNotFound(t *testing.T) {
	handler := NewUserHandler(&mockUserService{
		updateProfileFn: func(_ context.Context, _ int64, _ service.UpdateProfileInput) (*models.User, error) {
			return nil, service.ErrUserNotFound
		},
	})

	router := setupUserRouter(handler)
	body := `{"display_name":"New"}`
	req := httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
