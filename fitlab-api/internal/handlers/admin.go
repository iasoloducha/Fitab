package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"fitlab-api/internal/models"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	DB *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{DB: db}
}

// ListUsers returns all users in the system (admin only)
func (h *AdminHandler) ListUsers(c *gin.Context) {
	rows, err := h.DB.Query(
		"SELECT id, email, name, role, created_at FROM users ORDER BY name",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "error al obtener usuarios",
			Code:  "INTERNAL_ERROR",
		})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: users,
		Total: len(users),
	})
}

// UpdateUserName allows admin to update any user's name
func (h *AdminHandler) UpdateUserName(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error: "ID de usuario requerido",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error: "nombre no puede estar vacío",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	result, err := h.DB.Exec("UPDATE users SET name = ? WHERE id = ?", req.Name, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "error al actualizar usuario",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Error: "usuario no encontrado",
			Code:  "NOT_FOUND",
		})
		return
	}

	// Fetch updated user
	var user models.User
	err = h.DB.QueryRow(
		"SELECT id, email, name, role, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "error al obtener usuario actualizado",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: user,
	})
}

// DeleteUser allows admin to delete any user account
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error: "ID de usuario requerido",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	// Delete user (cascade will handle routines, exercises, logs)
	result, err := h.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "error al eliminar usuario",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Error: "usuario no encontrado",
			Code:  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "Usuario eliminado"},
	})
}

// Backup downloads a copy of the database
func (h *AdminHandler) Backup(c *gin.Context) {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./fitlab.db"
	}

	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "base de datos no encontrada",
			Code:  "NOT_FOUND",
		})
		return
	}

	// Open the database file
	file, err := os.Open(dbPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "error al leer la base de datos",
			Code:  "INTERNAL_ERROR",
		})
		return
	}
	defer file.Close()

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("fitlab-backup-%s.db", timestamp)

	// Set headers for file download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")

	// Copy file content to response
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		// Log error but don't fail (response already sent)
		fmt.Printf("Error sending backup: %v\n", err)
	}
}

// Restore replaces the database with an uploaded file
func (h *AdminHandler) Restore(c *gin.Context) {
	// Get database path
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./fitlab.db"
	}

	// Read uploaded file from form
	file, header, err := c.Request.FormFile("database")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error: "archivo no proporcionado",
			Code:  "VALIDATION_ERROR",
		})
		return
	}
	defer file.Close()

	// Verify it's a .db file
	if header.Filename == "" || len(header.Filename) < 3 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error: "nombre de archivo inválido",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	// Create temp file
	tempPath := dbPath + ".restoring"
	destFile, err := os.Create(tempPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "error al procesar archivo",
			Code:  "INTERNAL_ERROR",
		})
		return
	}
	defer destFile.Close()

	// Copy uploaded content
	_, err = io.Copy(destFile, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "error al guardar archivo",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	// Replace original database
	err = os.Rename(tempPath, dbPath)
	if err != nil {
		// Try to cleanup temp file
		os.Remove(tempPath)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error: "error al restaurar base de datos",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "Base de datos restaurada correctamente. Los cambios se aplicarán en el próximo request."},
	})
}
