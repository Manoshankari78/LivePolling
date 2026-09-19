package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"live-polling-app/backend/controllers"
	"live-polling-app/backend/middleware"
	"live-polling-app/backend/utils"
	pollws "live-polling-app/backend/websocket"
)

func Register(r *gin.Engine, auth *controllers.AuthController, polls *controllers.PollController, jwt *utils.JWTManager, hub *pollws.Hub, voteLimiter *middleware.RateLimiter) {
	api := r.Group("/api")
	api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	authRoutes := api.Group("/auth")
	authRoutes.POST("/register", auth.Register)
	authRoutes.POST("/login", auth.Login)
	authRoutes.GET("/me", middleware.AuthRequired(jwt), auth.Me)

	pollRoutes := api.Group("/polls")
	pollRoutes.POST("", middleware.AuthRequired(jwt), polls.Create)
	pollRoutes.GET("", middleware.AuthRequired(jwt), polls.List)
	pollRoutes.GET("/:id", polls.Get)
	pollRoutes.PUT("/:id", middleware.AuthRequired(jwt), polls.Update)
	pollRoutes.DELETE("/:id", middleware.AuthRequired(jwt), polls.Delete)
	pollRoutes.POST("/:id/close", middleware.AuthRequired(jwt), polls.Close)
	pollRoutes.POST("/:id/vote", voteLimiter.Middleware(), polls.Vote)
	pollRoutes.GET("/:id/results", polls.Results)
	pollRoutes.GET("/:id/ws", func(c *gin.Context) { hub.ServeHTTP(c.Writer, c.Request, c.Param("id")) })
}
