package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"fitlab-api/internal/middleware"
	"fitlab-api/internal/models"

	"github.com/gin-gonic/gin"
)

type RoutineHandler struct {
	DB *sql.DB
}

func NewRoutineHandler(db *sql.DB) *RoutineHandler {
	return &RoutineHandler{DB: db}
}

// List returns all routines for the current user
// Students see their own, professors see routines they created or their students' routines
func (h *RoutineHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var rows *sql.Rows
	var err error

	if role == "professor" {
		// Professors see routines they created
		rows, err = h.DB.Query(`
			SELECT r.id, r.user_id, r.created_by, r.title, r.start_date, r.end_date, r.is_active, r.created_at,
			       u.name as student_name
			FROM routines r
			JOIN users u ON r.user_id = u.id
			WHERE r.created_by = ? OR r.created_by IS NULL
			ORDER BY r.created_at DESC`,
			userID)
	} else {
		// Students see their own routines
		rows, err = h.DB.Query(`
			SELECT r.id, r.user_id, r.created_by, r.title, r.start_date, r.end_date, r.is_active, r.created_at,
			       u.name as student_name
			FROM routines r
			JOIN users u ON r.user_id = u.id
			WHERE r.user_id = ?
			ORDER BY r.created_at DESC`,
			userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al obtener rutinas",
			"code":  "INTERNAL_ERROR",
		})
		return
	}
	defer rows.Close()

	var routines []gin.H
	for rows.Next() {
		var r models.Routine
		var createdBy sql.NullInt64
		var startDate, endDate sql.NullString
		var studentName string

		if err := rows.Scan(&r.ID, &r.UserID, &createdBy, &r.Title, &startDate, &endDate, &r.IsActive, &r.CreatedAt, &studentName); err != nil {
			continue
		}

		if createdBy.Valid {
			r.CreatedBy = createdBy.Int64
		}
		if startDate.Valid {
			r.StartDate = startDate.String
		}
		if endDate.Valid {
			r.EndDate = endDate.String
		}

		routines = append(routines, gin.H{
			"id":            r.ID,
			"user_id":       r.UserID,
			"created_by":     r.CreatedBy,
			"title":        r.Title,
			"start_date":    r.StartDate,
			"end_date":     r.EndDate,
			"is_active":    r.IsActive,
			"created_at":   r.CreatedAt,
			"student_name": studentName,
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data:  routines,
		Total: len(routines),
	})
}

// Get returns a single routine with all exercises
func (h *RoutineHandler) Get(c *gin.Context) {
	routineID := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var r models.Routine
	var createdBy sql.NullInt64
	var startDate, endDate sql.NullString

	err := h.DB.QueryRow(`
		SELECT id, user_id, created_by, title, start_date, end_date, is_active, created_at
		FROM routines WHERE id = ?`,
		routineID,
	).Scan(&r.ID, &r.UserID, &createdBy, &r.Title, &startDate, &endDate, &r.IsActive, &r.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "rutina no encontrada",
			"code":  "NOT_FOUND",
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

	// Authorization: owner or creator (professor) can view
	if role == "student" && r.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "no tienes acceso a esta rutina",
			"code":  "FORBIDDEN",
		})
		return
	}

	if createdBy.Valid {
		r.CreatedBy = createdBy.Int64
	}
	if startDate.Valid {
		r.StartDate = startDate.String
	}
	if endDate.Valid {
		r.EndDate = endDate.String
	}

	// Get exercises
	exerciseRows, err := h.DB.Query(`
		SELECT e.id, e.routine_id, e.day_number, e.name, e.sets, e.reps, e.weight_kg, e.observations, e.sort_order, e.created_at, COALESCE(ec.image_urls, '')
		FROM exercises e
		LEFT JOIN exercise_catalog ec ON e.catalog_exercise_id = ec.id
		WHERE e.routine_id = ?
		ORDER BY e.day_number, e.sort_order`,
		r.ID,
	)
	if err == nil {
		defer exerciseRows.Close()
		for exerciseRows.Next() {
			var e models.Exercise
			var weightKg sql.NullString
			var observations sql.NullString
			var imageURLs sql.NullString

			if err := exerciseRows.Scan(&e.ID, &e.RoutineID, &e.DayNumber, &e.Name, &e.Sets, &e.Reps, &weightKg, &observations, &e.SortOrder, &e.CreatedAt, &imageURLs); err != nil {
				continue
			}
			if weightKg.Valid {
				e.WeightKg = weightKg.String
			}
			if observations.Valid {
				e.Observations = observations.String
			}
			if imageURLs.Valid && imageURLs.String != "" {
				e.ImageURLs = &imageURLs.String
			}
			r.Exercises = append(r.Exercises, e)
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: r,
	})
}

// Create creates a new routine with default exercises (3 days, 6 exercises each)
func (h *RoutineHandler) Create(c *gin.Context) {
	var req models.CreateRoutineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	userID := middleware.GetUserID(c)

	// Professors can assign to any student, students create for themselves
	if middleware.GetUserRole(c) == "student" {
		req.UserID = userID
	}

	// Default exercises: 3 days, 6 exercises each
	// "s" suffix indicates the JSON field is a string type in the request
	defaultExercises := []struct {
		DayNumber int
		Name      string
		Sets      int
		Reps      string
	}{
		// Day 1: Chest & Triceps
		{1, "Press de banca plano", 4, "8-10"},
		{1, "Press de banca inclinado", 3, "10-12"},
		{1, "Aperturas con mancuernas", 3, "12"},
		{1, "Fondos en paralelas", 3, "10-12"},
		{1, "Extensiones de triceps polea", 3, "12-15"},
		{1, "Extensiones de triceps cabeza", 3, "12-15"},
		// Day 2: Back & Biceps
		{2, "Dominadas o Jalón al pecho", 4, "8-10"},
		{2, "Remo con barra", 4, "8-10"},
		{2, "Remo con mancuernas", 3, "10-12"},
		{2, "Curl con barra", 3, "10-12"},
		{2, "Curl con mancuernas alternado", 3, "10-12 c/p"},
		{2, "Curl martillo", 3, "12"},
		// Day 3: Legs
		{3, "Sentadillas", 4, "8-10"},
		{3, "Prensa de piernas", 3, "10-12"},
		{3, "Extensiones de cuádriceps", 3, "12-15"},
		{3, "Curl de femoral", 3, "12-15"},
		{3, "Elevación de talones", 4, "15-20"},
		{3, "Abducción de cadera", 3, "15-20"},
	}

	// Use transaction to ensure all-or-nothing
	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al crear rutina",
			"code":  "INTERNAL_ERROR",
		})
		return
	}
	defer tx.Rollback()

	// Create routine
	result, err := tx.Exec(`
		INSERT INTO routines (user_id, created_by, title, start_date, end_date)
		VALUES (?, ?, ?, ?, ?)`,
		req.UserID, userID, req.Title, nullString(req.StartDate), nullString(req.EndDate),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al crear rutina",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	routineID, _ := result.LastInsertId()

	// Insert default exercises
	for i, ex := range defaultExercises {
		sortOrder := (ex.DayNumber-1)*6 + (i%6) + 1
		_, err = tx.Exec(`
			INSERT INTO exercises (routine_id, day_number, name, sets, reps, weight_kg, observations, sort_order)
			VALUES (?, ?, ?, ?, ?, '', '', ?)`,
			routineID, ex.DayNumber, ex.Name, ex.Sets, ex.Reps, sortOrder,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error al crear ejercicios por defecto",
				"code":  "INTERNAL_ERROR",
			})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error al crear rutina",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Data: gin.H{"id": routineID, "message": "rutina creada con ejercicios por defecto"},
	})
}

// Update updates a routine
func (h *RoutineHandler) Update(c *gin.Context) {
	routineID := c.Param("id")
	userID := middleware.GetUserID(c)

	// Verify ownership
	var ownerID int64
	err := h.DB.QueryRow("SELECT user_id FROM routines WHERE id = ?", routineID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "rutina no encontrada"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	// Allow update if: student is owner OR any professor (can change is_active)
	userRole := middleware.GetUserRole(c)
	isOwner := ownerID == userID
	
	if userRole == "student" && !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "no tienes permiso"})
		return
	}

	var req models.UpdateRoutineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
		return
	}

	_, err = h.DB.Exec(`
		UPDATE routines SET title = COALESCE(NULLIF(?, ''), title),
		                start_date = COALESCE(NULLIF(?, ''), start_date),
		                end_date = COALESCE(NULLIF(?, ''), end_date),
		                is_active = COALESCE(?, is_active)
		WHERE id = ?`,
		req.Title, req.StartDate, req.EndDate, req.IsActive, routineID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "rutina actualizada"},
	})
}

// Delete removes a routine
func (h *RoutineHandler) Delete(c *gin.Context) {
	routineID := c.Param("id")
	userID := middleware.GetUserID(c)

	if middleware.GetUserRole(c) == "student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "solo profesores pueden eliminar rutinas"})
		return
	}

	result, err := h.DB.Exec("DELETE FROM routines WHERE id = ? AND created_by = ?", routineID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "rutina no encontrada o sin permisos"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "rutina eliminada"},
	})
}

// Copy creates a copy of an existing routine for another user
func (h *RoutineHandler) Copy(c *gin.Context) {
	routineID := c.Param("id")

	// Verify user is a professor
	if middleware.GetUserRole(c) != "professor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "solo profesores pueden copiar rutinas"})
		return
	}

	var req models.CopyRoutineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	// Verify source routine exists
	var sourceRoutine models.Routine
	var sourceCreatedBy sql.NullInt64
	var sourceStartDate, sourceEndDate sql.NullString
	var sourceTitle string
	var sourceIsActive bool

	err := h.DB.QueryRow(`
		SELECT id, user_id, created_by, title, start_date, end_date, is_active
		FROM routines WHERE id = ?`,
		routineID,
	).Scan(&sourceRoutine.ID, &sourceRoutine.UserID, &sourceCreatedBy, &sourceTitle, &sourceStartDate, &sourceEndDate, &sourceIsActive)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "rutina no encontrada"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	// Use provided title or default to source title + "(copia)"
	title := req.Title
	if title == "" {
		title = sourceTitle + " (copia)"
	}

	// Verify target user exists and is a student
	var targetUserRole string
	err = h.DB.QueryRow("SELECT role FROM users WHERE id = ?", req.TargetUserID).Scan(&targetUserRole)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alumno no encontrado"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	if targetUserRole != "student" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "solo se puede copiar a alumnos"})
		return
	}

	// Get source exercises
	exerciseRows, err := h.DB.Query(`
		SELECT day_number, name, sets, reps, weight_kg, observations, sort_order
		FROM exercises WHERE routine_id = ?
		ORDER BY day_number, sort_order`,
		routineID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al obtener ejercicios"})
		return
	}
	defer exerciseRows.Close()

	var exercises []struct {
		DayNumber    int
		Name         string
		Sets         int
		Reps         string
		WeightKg    string
		Observations string
		SortOrder    int
	}

	for exerciseRows.Next() {
		var e struct {
			DayNumber    int
			Name         string
			Sets         int
			Reps         string
			WeightKg    string
			Observations string
			SortOrder    int
		}
		var weightKg, observations sql.NullString
		if err := exerciseRows.Scan(&e.DayNumber, &e.Name, &e.Sets, &e.Reps, &weightKg, &observations, &e.SortOrder); err != nil {
			continue
		}
		if weightKg.Valid {
			e.WeightKg = weightKg.String
		}
		if observations.Valid {
			e.Observations = observations.String
		}
		exercises = append(exercises, e)
	}

	// Create new routine with exercises in a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al copiar rutina"})
		return
	}
	defer tx.Rollback()

	userID := middleware.GetUserID(c)
	result, err := tx.Exec(`
		INSERT INTO routines (user_id, created_by, title, start_date, end_date, is_active)
		VALUES (?, ?, ?, ?, ?, ?)`,
		req.TargetUserID, userID, title,
		nullString(sourceStartDate.String),
		nullString(sourceEndDate.String),
		sourceIsActive,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al copiar rutina"})
		return
	}

	newRoutineID, _ := result.LastInsertId()

	// Copy all exercises
	for _, ex := range exercises {
		_, err = tx.Exec(`
			INSERT INTO exercises (routine_id, day_number, name, sets, reps, weight_kg, observations, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			newRoutineID, ex.DayNumber, ex.Name, ex.Sets, ex.Reps, ex.WeightKg, ex.Observations, ex.SortOrder,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error al copiar ejercicios"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al copiar rutina"})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Data: gin.H{"id": newRoutineID, "message": "rutina copiada exitosamente"},
	})
}

// Helper: convert empty string to nil
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// Helper: parse int64 from string
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}