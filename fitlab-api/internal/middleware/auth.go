package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

const (
	SessionName    = "fitlab-session"
	SessionUserID  = "user_id"
	SessionUserRole = "user_role"
	SessionEmail   = "user_email"
)

// AuthMiddleware requires a valid session
func AuthMiddleware(store sessions.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := store.Get(c.Request, SessionName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "sesión inválida",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		userID, ok := session.Values[SessionUserID].(int64)
		if !ok || userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "debes iniciar sesión",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Store user info in context for handlers
		c.Set(SessionUserID, userID)
		c.Set(SessionUserRole, session.Values[SessionUserRole].(string))
		c.Set(SessionEmail, session.Values[SessionEmail].(string))

		c.Next()
	}
}

// ProfessorOnly checks if the user is a professor
func ProfessorOnly(c *gin.Context) {
	role, exists := c.Get(SessionUserRole)
	if !exists || role != "professor" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "solo profesores pueden acceder",
			"code":  "FORBIDDEN",
		})
		return
	}
	c.Next()
}

// GetUserID returns the user ID from context
func GetUserID(c *gin.Context) int64 {
	id, exists := c.Get(SessionUserID)
	if !exists {
		return 0
	}
	v, ok := id.(int64)
	if !ok {
		return 0
	}
	return v
}

// GetUserRole returns the user role from context
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get(SessionUserRole)
	if !exists {
		return ""
	}
	v, ok := role.(string)
	if !ok {
		return ""
	}
	return v
}