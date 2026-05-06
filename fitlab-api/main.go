package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strings"

	"fitlab-api/internal/database"
	"fitlab-api/internal/handlers"
	"fitlab-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func main() {
	// Configuration
	dbPath := getEnv("DATABASE_PATH", "./fitlab.db")
	// Fly.io provides PORT env var; fallback to ADDR then default :8080
	port := getEnv("PORT", "")
	addr := getEnv("ADDR", ":8080")
	if port != "" {
		addr = ":" + port
	}
	sessionSecret := getEnv("SESSION_SECRET", generateRandomSecret())
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:8080")
	isDevMode := getEnv("DEV_MODE", "true") == "true"

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

    //Samesite dinamico segun el ambiente
	sameSite := http.SameSiteLaxMode
	if !isDevMode {
		sameSite = http.SameSiteNoneMode
	}
	
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   !isDevMode, // true in production (HTTPS), false in development
		SameSite: sameSite,
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db.DB, sessionStore)
	adminHandler := handlers.NewAdminHandler(db.DB)
	routineHandler := handlers.NewRoutineHandler(db.DB)
	exerciseHandler := handlers.NewExerciseHandler(db.DB)

	// Setup router
	r := gin.Default()

	// CORS for frontend development
	r.Use(corsMiddleware(allowedOrigins))

	// API routes
	api := r.Group("/api")
	{
		// Public auth routes
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/forgot-password", authHandler.ForgotPassword)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(sessionStore))
		{
			// Auth
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.Me)
			protected.PUT("/auth/password", authHandler.ChangePassword)

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

	// Admin routes (public login)
	api.POST("/admin/login", authHandler.AdminLogin)

	// Admin protected routes
	admin := api.Group("/admin")
	admin.Use(middleware.AdminMiddleware(sessionStore))
	{
		admin.GET("/users", adminHandler.ListUsers)
		admin.PUT("/users/:id", adminHandler.UpdateUserName)
 		admin.DELETE("/users/:id", adminHandler.DeleteUser)
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

// corsMiddleware validates origin against whitelist
func corsMiddleware(allowedOrigins string) gin.HandlerFunc {
	// Parse allowed origins into a map for O(1) lookup
	originMap := make(map[string]bool)
	for _, o := range splitAndTrim(allowedOrigins, ",") {
		originMap[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is in whitelist
		if originMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			// Still allow preflight even if origin failed (browser needs the error response)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// splitAndTrim splits a string by delimiter and trims whitespace
func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// getEnv returns env var or default
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// generateRandomSecret generates a random 32-byte secret for sessions
func generateRandomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate random secret: %v", err)
	}
	return hex.EncodeToString(b)
}