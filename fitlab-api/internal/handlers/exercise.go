package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"fitlab-api/internal/middleware"
	"fitlab-api/internal/models"

	"github.com/gin-gonic/gin"
)

type ExerciseHandler struct {
	DB *sql.DB
}

func NewExerciseHandler(db *sql.DB) *ExerciseHandler {
	return &ExerciseHandler{DB: db}
}

// Create adds an exercise to a routine
func (h *ExerciseHandler) Create(c *gin.Context) {
	routineID := c.Param("id")

	// Verify professor owns this routine
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

	// Authorization
	if middleware.GetUserRole(c) == "student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "solo profesores pueden agregar ejercicios"})
		return
	}

	var req models.CreateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos: " + err.Error()})
		return
	}

	sortOrder := req.SortOrder
	if sortOrder == 0 {
		// Get next sort order for this day
		var maxOrder sql.NullInt64
		h.DB.QueryRow("SELECT MAX(sort_order) FROM exercises WHERE routine_id = ? AND day_number = ?", routineID, req.DayNumber).Scan(&maxOrder)
		if maxOrder.Valid {
			sortOrder = int(maxOrder.Int64) + 1
		}
	}

	result, err := h.DB.Exec(`
		INSERT INTO exercises (routine_id, day_number, name, sets, reps, weight_kg, observations, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		routineID, req.DayNumber, req.Name, req.Sets, req.Reps, req.WeightKg, req.Observations, sortOrder,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al crear ejercicio"})
		return
	}

	exerciseID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, models.APIResponse{
		Data: gin.H{"id": exerciseID, "message": "ejercicio agregado"},
	})
}

// Update updates an exercise
func (h *ExerciseHandler) Update(c *gin.Context) {
	exerciseID := c.Param("id")
	userID := middleware.GetUserID(c)

	// Verify ownership through routine
	var routineID int64
	var professorID int64
	err := h.DB.QueryRow(`
		SELECT r.id, r.created_by FROM exercises e
		JOIN routines r ON e.routine_id = r.id
		WHERE e.id = ?`, exerciseID,
	).Scan(&routineID, &professorID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "ejercicio no encontrado"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	if middleware.GetUserRole(c) == "student" || professorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "sin permiso"})
		return
	}

	var req models.UpdateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
		return
	}

	_, err = h.DB.Exec(`
		UPDATE exercises SET
			day_number = COALESCE(?, day_number),
			name = COALESCE(NULLIF(?, ''), name),
			sets = COALESCE(?, sets),
			reps = COALESCE(NULLIF(?, ''), reps),
			weight_kg = ?,
			observations = COALESCE(NULLIF(?, ''), observations),
			sort_order = COALESCE(?, sort_order)
		WHERE id = ?`,
		req.DayNumber, req.Name, req.Sets, req.Reps, req.WeightKg, req.Observations, req.SortOrder, exerciseID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "ejercicio actualizado"},
	})
}

// Delete removes an exercise
func (h *ExerciseHandler) Delete(c *gin.Context) {
	exerciseID := c.Param("id")
	userID := middleware.GetUserID(c)

	if middleware.GetUserRole(c) == "student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "solo profesores pueden eliminar ejercicios"})
		return
	}

	result, err := h.DB.Exec(`
		DELETE FROM exercises WHERE id = ? AND routine_id IN (
			SELECT routine_id FROM exercises WHERE id = ? AND routine_id IN (
				SELECT id FROM routines WHERE created_by = ?
			)
		)`,
		exerciseID, exerciseID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ejercicio no encontrado o sin permisos"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "ejercicio eliminado"},
	})
}

// LogCompletion logs a student's completion for an exercise
func (h *ExerciseHandler) LogCompletion(c *gin.Context) {
	exerciseID := c.Param("id")
	userID := middleware.GetUserID(c)

	// Verify exercise exists and belongs to user's routine
	var exists bool
	err := h.DB.QueryRow(`
		SELECT 1 FROM exercises e
		JOIN routines r ON e.routine_id = r.id
		WHERE e.id = ? AND r.user_id = ?`, exerciseID, userID,
	).Scan(&exists)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusForbidden, gin.H{"error": "este ejercicio no pertenece a tus rutinas"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	var req models.LogExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
		return
	}

	result, err := h.DB.Exec(`
		INSERT INTO exercise_logs (exercise_id, user_id, date, completed, actual_weight, notes)
		VALUES (?, ?, ?, ?, ?, ?)`,
		exerciseID, userID, req.Date, req.Completed, req.ActualWeight, req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al guardar"})
		return
	}

	logID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, models.APIResponse{
		Data: gin.H{"id": logID, "message": "progreso registrado"},
	})
}

// GetLogs returns all logs for an exercise
func (h *ExerciseHandler) GetLogs(c *gin.Context) {
	exerciseID := c.Param("id")
	userID := middleware.GetUserID(c)

	rows, err := h.DB.Query(`
		SELECT id, exercise_id, date, completed, actual_weight, notes, created_at
		FROM exercise_logs
		WHERE exercise_id = ? AND user_id = ?
		ORDER BY date DESC`,
		exerciseID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al obtener logs"})
		return
	}
	defer rows.Close()

	var logs []models.ExerciseLog
	for rows.Next() {
		var log models.ExerciseLog
		var weight sql.NullString
		var notes sql.NullString

		if err := rows.Scan(&log.ID, &log.ExerciseID, &log.Date, &log.Completed, &weight, &notes, &log.CreatedAt); err != nil {
			continue
		}
		if weight.Valid {
			log.ActualWeight = weight.String
		}
		if notes.Valid {
			log.Notes = notes.String
		}
		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data:  logs,
		Total: len(logs),
	})
}

// DeleteLog deletes a specific exercise log entry
func (h *ExerciseHandler) DeleteLog(c *gin.Context) {
	logID := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	// Verify the log belongs to the user (or professor can delete any from their students)
	if role == "student" {
		var count int
		err := h.DB.QueryRow(`
			SELECT COUNT(*) FROM exercise_logs 
			WHERE id = ? AND user_id = ?`,
			logID, userID,
		).Scan(&count)
		
		if err != nil || count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "no tienes permiso para eliminar este registro"})
			return
		}
	} else {
		// Professor can delete logs from their students
		var count int
		err := h.DB.QueryRow(`
			SELECT COUNT(*) FROM exercise_logs el
			JOIN routines r ON el.user_id = r.user_id
			WHERE el.id = ? AND r.created_by = ?`,
			logID, userID,
		).Scan(&count)
		
		if err != nil || count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "no tienes permiso para eliminar este registro"})
			return
		}
	}

	result, err := h.DB.Exec("DELETE FROM exercise_logs WHERE id = ?", logID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data: gin.H{"message": "registro eliminado"},
	})
}

// GetProgress returns a student's progress across all exercises
func (h *ExerciseHandler) GetProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	// If professor and user_id query param provided, get that user's progress
	if role == "professor" {
		if targetUserID := c.Query("user_id"); targetUserID != "" {
			// Verify professor created routines for this user
			var count int
			h.DB.QueryRow("SELECT COUNT(*) FROM routines WHERE created_by = ? AND user_id = ?", userID, targetUserID).Scan(&count)
			if count == 0 {
				c.JSON(http.StatusForbidden, gin.H{"error": "no tienes permiso para ver el progreso de este alumno"})
				return
			}
			userID, _ = strconv.ParseInt(targetUserID, 10, 64)
		}
	}

	from := c.Query("from")
	to := c.Query("to")

	query := `
		SELECT el.id, el.exercise_id, el.date, el.completed, el.actual_weight, el.notes, el.created_at,
		       e.name as exercise_name
		FROM exercise_logs el
		JOIN exercises e ON el.exercise_id = e.id
		JOIN routines r ON e.routine_id = r.id
		WHERE el.user_id = ? AND r.user_id = ?`
	args := []interface{}{userID, userID}

	if from != "" {
		query += " AND el.date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND el.date <= ?"
		args = append(args, to)
	}

	query += " ORDER BY el.date DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al obtener progreso"})
		return
	}
	defer rows.Close()

	var progress []gin.H
	for rows.Next() {
		var log models.ExerciseLog
		var exerciseName string
		var weight sql.NullString
		var notes sql.NullString

		if err := rows.Scan(&log.ID, &log.ExerciseID, &log.Date, &log.Completed, &weight, &notes, &log.CreatedAt, &exerciseName); err != nil {
			continue
		}
		if weight.Valid {
			log.ActualWeight = weight.String
		}
		if notes.Valid {
			log.Notes = notes.String
		}

		progress = append(progress, gin.H{
			"id":            log.ID,
			"exercise_id":   log.ExerciseID,
			"exercise_name": exerciseName,
			"date":         log.Date,
			"completed":    log.Completed,
			"actual_weight": log.ActualWeight,
			"notes":        log.Notes,
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Data:  progress,
		Total: len(progress),
	})
}

// Helper: parse int from context param
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}