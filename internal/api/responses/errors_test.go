package responses

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWriteError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		WriteError(c, "health_check_failed", "Health status unknown", http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"health_check_failed"`)
	assert.Contains(t, w.Body.String(), `"message":"Health status unknown"`)
	assert.Contains(t, w.Body.String(), `"timestamp"`)
}

func TestWriteErrorDefaultStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		WriteError(c, "generic_error", "Something went wrong", 0)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestErrorResponseTimestampUsesUTC(t *testing.T) {
	response := ErrorResponse{}
	response.Error.Timestamp = time.Now().UTC()

	assert.Equal(t, time.UTC, response.Error.Timestamp.Location())
}
