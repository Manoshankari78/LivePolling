package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"live-polling-app/backend/middleware"
	"live-polling-app/backend/services"
)

type PollController struct{ service *services.PollService }

func NewPollController(service *services.PollService) *PollController {
	return &PollController{service: service}
}

type pollRequest struct {
	Question  string     `json:"question"`
	Options   []string   `json:"options"`
	ExpiresAt *time.Time `json:"expiresAt"`
}
type voteRequest struct {
	OptionID string `json:"optionId"`
	VoterID  string `json:"voterId"`
}

func parseID(c *gin.Context) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return primitive.NilObjectID, false
	}
	return id, true
}
func (h *PollController) Create(c *gin.Context) {
	creator, err := primitive.ObjectIDFromHex(middleware.CurrentUserID(c))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	var req pollRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	poll, err := h.service.Create(c.Request.Context(), creator, req.Question, req.Options, req.ExpiresAt)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"poll": poll})
}
func (h *PollController) List(c *gin.Context) {
	creator, err := primitive.ObjectIDFromHex(middleware.CurrentUserID(c))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	polls, err := h.service.List(c.Request.Context(), creator)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"polls": polls})
}
func (h *PollController) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	poll, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"poll": poll})
}
func (h *PollController) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	creator, err := primitive.ObjectIDFromHex(middleware.CurrentUserID(c))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	var req pollRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	poll, err := h.service.Update(c.Request.Context(), id, creator, req.Question, req.Options, req.ExpiresAt)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"poll": poll})
}
func (h *PollController) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	creator, err := primitive.ObjectIDFromHex(middleware.CurrentUserID(c))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	if err = h.service.Delete(c.Request.Context(), id, creator); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *PollController) Close(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	creator, err := primitive.ObjectIDFromHex(middleware.CurrentUserID(c))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	if err = h.service.Close(c.Request.Context(), id, creator); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "poll closed"})
}
func (h *PollController) Vote(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req voteRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	results, err := h.service.Vote(c.Request.Context(), id, req.OptionID, req.VoterID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
func (h *PollController) Results(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	results, err := h.service.Results(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
