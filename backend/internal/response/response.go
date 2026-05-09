package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:      0,
		Message:   "ok",
		Data:      data,
		RequestID: requestID(c),
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Body{
		Code:      code,
		Message:   message,
		RequestID: requestID(c),
	})
}

func requestID(c *gin.Context) string {
	value, exists := c.Get("request_id")
	if !exists {
		return ""
	}

	requestID, ok := value.(string)
	if !ok {
		return ""
	}

	return requestID
}
