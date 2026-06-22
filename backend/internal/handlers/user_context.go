package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const defaultUserID uint = 1

func userIDFromRequest(c *gin.Context) uint {
	value, exists := c.Get("user_id")
	if !exists {
		return defaultUserID
	}

	switch typed := value.(type) {
	case uint:
		if typed > 0 {
			return typed
		}
	case uint64:
		if typed > 0 {
			return uint(typed)
		}
	case int:
		if typed > 0 {
			return uint(typed)
		}
	}

	return defaultUserID
}

func uintParam(c *gin.Context, name string) (uint, error) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func uintFromString(raw string) uint {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0
	}
	return uint(value)
}
