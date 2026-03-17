package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddlewarePreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), http.MethodGet)
}

func TestIsAllowedOrigin(t *testing.T) {
	testCases := []struct {
		name     string
		origin   string
		expected bool
	}{
		{name: "localhost 3000 allowed", origin: "http://localhost:3000", expected: true},
		{name: "localhost 8080 allowed", origin: "http://localhost:8080", expected: true},
		{name: "loopback any port allowed", origin: "http://127.0.0.1:9999", expected: true},
		{name: "https localhost not allowed", origin: "https://localhost:3000", expected: false},
		{name: "localhost wrong port denied", origin: "http://localhost:4000", expected: false},
		{name: "spoofed prefix denied", origin: "http://127.0.0.1.evil.com:3000", expected: false},
		{name: "empty denied", origin: "", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, isAllowedOrigin(testCase.origin))
		})
	}
}
