package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"live-polling-app/backend/config"
	"live-polling-app/backend/controllers"
	"live-polling-app/backend/database"
	"live-polling-app/backend/middleware"
	"live-polling-app/backend/repositories"
	"live-polling-app/backend/routes"
	"live-polling-app/backend/services"
	"live-polling-app/backend/utils"
	pollws "live-polling-app/backend/websocket"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "dev-only-change-me" {
		log.Println("WARNING: using development JWT secret; set JWT_SECRET for real deployment")
	}

	mongoClient, db, err := database.ConnectMongo(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatalf("mongo connection failed: %v", err)
	}
	defer func() { _ = mongoClient.Disconnect(context.Background()) }()

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis URL invalid: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer rdb.Close()

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpires)
	userRepo := repositories.NewUserRepository(db)
	pollRepo := repositories.NewPollRepository(db)
	voteRepo := repositories.NewVoteRepository(db)
	authService := services.NewAuthService(userRepo, jwtManager)
	pollService := services.NewPollService(pollRepo, voteRepo)
	authController := controllers.NewAuthController(authService)
	pollController := controllers.NewPollController(pollService)
	hub := pollws.NewHub(cfg.CORSOrigin)
	pollws.StartRedisSubscriber(rdb, hub)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	corsCfg := cors.Config{AllowOrigins: split(cfg.CORSOrigin), AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Voter-ID"}, AllowCredentials: true, MaxAge: 12 * time.Hour}
	r.Use(cors.New(corsCfg))
	r.MaxMultipartMemory = 8 << 20
	routes.Register(r, authController, pollController, jwtManager, hub, middleware.NewRateLimiter(30, time.Minute))

	server := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("live polling backend listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && !strings.Contains(err.Error(), "Server closed") {
		log.Fatal(err)
	}
}
func split(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
