package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"live-polling-app/backend/services"
)

func handleError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, services.ErrBadRequest):
		status, message = http.StatusBadRequest, "invalid request"
	case errors.Is(err, services.ErrConflict):
		status, message = http.StatusConflict, "operation conflicts with the current resource state"
	case errors.Is(err, services.ErrUnauthorized):
		status, message = http.StatusUnauthorized, "authentication required"
	case errors.Is(err, services.ErrForbidden):
		status, message = http.StatusForbidden, "you are not allowed to modify this resource"
	case errors.Is(err, services.ErrNotFound), errors.Is(err, mongo.ErrNoDocuments):
		status, message = http.StatusNotFound, "resource not found"
	case errors.Is(err, services.ErrInvalidCredentials):
		status, message = http.StatusUnauthorized, "invalid email or password"
	case stringsEqual(err.Error(), "duplicate options are not allowed"):
		status, message = http.StatusBadRequest, err.Error()
	default:
		if err != nil && len(err.Error()) < 180 {
			status, message = http.StatusBadRequest, err.Error()
		}
	}
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
func stringsEqual(a, b string) bool { return a == b }
