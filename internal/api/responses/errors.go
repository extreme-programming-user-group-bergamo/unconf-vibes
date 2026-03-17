package responses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error struct {
		Code      string    `json:"code"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"error"`
}

func WriteError(c *gin.Context, code, message string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	response := ErrorResponse{}
	response.Error.Code = code
	response.Error.Message = message
	response.Error.Timestamp = time.Now().UTC()

	c.JSON(statusCode, response)
}
