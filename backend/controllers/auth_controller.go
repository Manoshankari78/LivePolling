package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"live-polling-app/backend/middleware"
	"live-polling-app/backend/services"
)

type AuthController struct{ service *services.AuthService }

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

type registerRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthController) Register(c *gin.Context) {
	var req registerRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
		return
	}
	user, token, err := h.service.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}
func (h *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	user, token, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}
func (h *AuthController) Me(c *gin.Context) {
	user, err := h.service.Me(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
