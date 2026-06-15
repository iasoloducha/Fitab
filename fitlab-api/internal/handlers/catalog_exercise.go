package handlers

import (
	"database/sql"
	"net/http"

	"fitlab-api/internal/models"

	"github.com/gin-gonic/gin"
)

type CatalogExerciseHandler struct {
	DB *sql.DB
}

func NewCatalogExerciseHandler(db *sql.DB) *CatalogExerciseHandler {
	return &CatalogExerciseHandler{DB: db}
}

// List returns all catalog exercises, optionally filtered by name
func (h *CatalogExerciseHandler) List(c *gin.Context) {
	q := c.Query("q")

	var rows *sql.Rows
	var err error

	if q != "" {
		rows, err = h.DB.Query(`
			SELECT id, name, COALESCE(image_urls, ''), created_at
			FROM exercise_catalog
			WHERE name LIKE '%' || ? || '%'
			ORDER BY name`,
			q)
	} else {
		rows, err = h.DB.Query(`
			SELECT id, name, COALESCE(image_urls, ''), created_at
			FROM exercise_catalog
			ORDER BY name`)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al obtener catálogo",
			"code":  "INTERNAL_ERROR",
		})
		return
	}
	defer rows.Close()

	var exercises []models.CatalogExercise
	for rows.Next() {
		var ex models.CatalogExercise
		if err := rows.Scan(&ex.ID, &ex.Name, &ex.ImageURLs, &ex.CreatedAt); err != nil {
			continue
		}
		exercises = append(exercises, ex)
	}

	if exercises == nil {
		exercises = []models.CatalogExercise{}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data:  exercises,
		Total: len(exercises),
	})
}

// Create adds a new exercise to the catalog
func (h *CatalogExerciseHandler) Create(c *gin.Context) {
	var req models.CreateCatalogExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos: " + err.Error(),
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	result, err := h.DB.Exec(`
		INSERT INTO exercise_catalog (name, image_urls)
		VALUES (?, ?)`,
		req.Name, nullString(req.ImageURLs),
	)
	if err != nil {
		// Check for UNIQUE constraint violation
		if sqlErr, ok := err.(interface{ Error() string }); ok && contains(sqlErr.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "ya existe un ejercicio con ese nombre",
				"code":  "DUPLICATE_NAME",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al crear ejercicio en el catálogo",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	exerciseID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, models.APIResponse{
		Data: gin.H{"id": exerciseID, "message": "ejercicio creado en el catálogo"},
	})
}

// Update modifies an existing catalog exercise
func (h *CatalogExerciseHandler) Update(c *gin.Context) {
	exerciseID := c.Param("id")

	var req models.UpdateCatalogExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	// Build dynamic UPDATE
	query := "UPDATE exercise_catalog SET "
	args := []interface{}{}
	parts := []string{}

	if req.Name != nil {
		parts = append(parts, "name = ?")
		args = append(args, *req.Name)
	}
	if req.ImageURLs != nil {
		parts = append(parts, "image_urls = ?")
		args = append(args, *req.ImageURLs)
	}

	if len(parts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no hay campos para actualizar",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	for i, p := range parts {
		if i > 0 {
			query += ", "
		}
		query += p
	}
	query += " WHERE id = ?"
	args = append(args, exerciseID)

	_, err := h.DB.Exec(query, args...)
	if err != nil {
		if contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "ya existe un ejercicio con ese nombre",
				"code":  "DUPLICATE_NAME",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al actualizar ejercicio",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "ejercicio actualizado"},
	})
}

// Delete removes a catalog exercise if it's not referenced by any routine exercise
func (h *CatalogExerciseHandler) Delete(c *gin.Context) {
	exerciseID := c.Param("id")

	// Check if any routine exercise references this catalog entry
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM exercises WHERE catalog_exercise_id = ?", exerciseID).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al verificar uso del ejercicio",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "el ejercicio está en uso en una o más rutinas",
			"code":  "IN_USE",
		})
		return
	}

	result, err := h.DB.Exec("DELETE FROM exercise_catalog WHERE id = ?", exerciseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al eliminar ejercicio",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "ejercicio no encontrado",
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "ejercicio eliminado del catálogo"},
	})
}

// contains is a simple string contains helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
