package handlers

import (
	"database/sql"
	"net/http"

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
