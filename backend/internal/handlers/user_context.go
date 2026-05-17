package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const defaultUserID uint = 1

func userIDFromRequest(c *gin.Context) uint {
	raw := c.GetHeader("X-User-ID")
	if raw == "" {
		return defaultUserID
	}

	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return defaultUserID
	}

	return uint(id)
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
