package handlers

import (
	"database/sql"
	"net/http"
	"os"

	"fitlab-api/internal/middleware"
	"fitlab-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

const devProfessorCode = "PROF2024" // ONLY for local development

type AuthHandler struct {
	DB    *sql.DB
	Store sessions.Store
}

func NewAuthHandler(db *sql.DB, store sessions.Store) *AuthHandler {
	return &AuthHandler{DB: db, Store: store}
}

// Register creates a new user account
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error: "datos inválidos: " + err.Error(),
			Code:  "VALIDATION_ERROR",
		})
		return
	}

// Professors need a valid registration code
	if req.Role == "professor" {
		expectedCode := os.Getenv("PROFESSOR_REGISTRATION_CODE")
		isDevMode := os.Getenv("DEV_MODE") == "true"

		// In development, use hardcoded fallback; in production, MUST be set
		if expectedCode == "" {
			if isDevMode {
				expectedCode = devProfessorCode // Only for local testing
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "PROFESSOR_REGISTRATION_CODE not configured",
					"code":  "CONFIG_MISSING",
				})
				return
			}
		}

		if req.ProfessorCode != expectedCode {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "código de profesor inválido",
				"code":  "INVALID_CODE",
			})
			return
		}
	}

	// Check if email already exists
	var exists int
	err := h.DB.QueryRow("SELECT 1 FROM users WHERE email = ?", req.Email).Scan(&exists)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "este email ya está registrado",
			"code":  "EMAIL_EXISTS",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al procesar contraseña",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	var user models.User
	result, err := h.DB.Exec(
		"INSERT INTO users (email, password_hash, name, role) VALUES (?, ?, ?, ?)",
		req.Email, string(hashedPassword), req.Name, req.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al crear usuario",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	userID, _ := result.LastInsertId()
	
	// Fetch complete user record
	err = h.DB.QueryRow(
		"SELECT id, email, name, role, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al obtener usuario",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Create session
	session, _ := h.Store.Get(c.Request, middleware.SessionName)
	session.Values[middleware.SessionUserID] = user.ID
	session.Values[middleware.SessionUserRole] = req.Role
	session.Values[middleware.SessionEmail] = req.Email
	session.Save(c.Request, c.Writer)

	c.JSON(http.StatusCreated, models.APIResponse{
		Data: user,
	})
}

// Login authenticates a user and creates a session
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error: "email y contraseña son requeridos",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	// Find user
	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, email, password_hash, name, role, created_at FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "email o contraseña incorrectos",
			"code":  "INVALID_CREDENTIALS",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error interno",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "email o contraseña incorrectos",
			"code":  "INVALID_CREDENTIALS",
		})
		return
	}

	// Create session
	session, _ := h.Store.Get(c.Request, middleware.SessionName)
	session.Values[middleware.SessionUserID] = user.ID
	session.Values[middleware.SessionUserRole] = user.Role
	session.Values[middleware.SessionEmail] = user.Email
	session.Save(c.Request, c.Writer)

	c.JSON(http.StatusOK, models.APIResponse{
		Data: user,
	})
}

// Logout destroys the session
func (h *AuthHandler) Logout(c *gin.Context) {
	session, _ := h.Store.Get(c.Request, middleware.SessionName)
	session.Options.MaxAge = -1
	session.Save(c.Request, c.Writer)

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "sesión cerrada"},
	})
}

// Me returns the current user info
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, email, name, role, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "usuario no encontrado",
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: user,
	})
}

// GetStudents returns all students (for professors to assign routines)
func (h *AuthHandler) GetStudents(c *gin.Context) {
	rows, err := h.DB.Query(
		"SELECT id, email, name, role, created_at FROM users WHERE role = 'student' ORDER BY name",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al obtener estudiantes",
			"code":  "INTERNAL_ERROR",
		})
		return
	}
	defer rows.Close()

	var students []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt); err != nil {
			continue
		}
		students = append(students, u)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data:  students,
		Total: len(students),
	})
}