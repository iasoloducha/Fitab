package main

import (
	"log"
	"net/http"
	"os"

	"fitlab-api/internal/database"
	"fitlab-api/internal/handlers"
	"fitlab-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func main() {
	// Configuration
	dbPath := getEnv("DATABASE_PATH", "./fitlab.db")
	addr := getEnv("ADDR", ":8080")
	sessionSecret := getEnv("SESSION_SECRET", "change-me-in-production-super-secret-key")

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Database initialized at %s", dbPath)

	// Initialize session store
	store := sessions.NewCookieStore([]byte(sessionSecret))
	var sessionStore sessions.Store = store
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // TODO: true in production
		SameSite: http.SameSiteLaxMode,
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db.DB, sessionStore)
	routineHandler := handlers.NewRoutineHandler(db.DB)
	exerciseHandler := handlers.NewExerciseHandler(db.DB)

	// Setup router
	r := gin.Default()

	// CORS for frontend development
	r.Use(corsMiddleware())

	// API routes
	api := r.Group("/api")
	{
		// Public auth routes
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(sessionStore))
		{
			// Auth
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.Me)

			// Routines
			protected.GET("/routines", routineHandler.List)
			protected.GET("/routines/:id", routineHandler.Get)

			// Professor-only routine management
			professor := protected.Group("")
			professor.Use(func(c *gin.Context) {
				middleware.AuthMiddleware(store)(c)
				if c.IsAborted() {
					return
				}
				middleware.ProfessorOnly(c)
			})
			{
				professor.POST("/routines", routineHandler.Create)
				professor.PUT("/routines/:id", routineHandler.Update)
				professor.DELETE("/routines/:id", routineHandler.Delete)
				professor.POST("/routines/:id/exercises", exerciseHandler.Create)
				professor.PUT("/exercises/:id", exerciseHandler.Update)
				professor.DELETE("/exercises/:id", exerciseHandler.Delete)
				professor.GET("/users/students", authHandler.GetStudents)
			}

			// Student exercise logging
			protected.POST("/exercises/:id/logs", exerciseHandler.LogCompletion)
			protected.GET("/exercises/:id/logs", exerciseHandler.GetLogs)
			protected.DELETE("/logs/:id", exerciseHandler.DeleteLog)
			protected.GET("/progress", exerciseHandler.GetProgress)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// corsMiddleware allows frontend to call the API
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getEnv returns env var or default
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}