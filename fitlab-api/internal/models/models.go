package models

import "time"

// User represents a user in the system (student or professor)
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         string    `json:"role"` // "student" or "professor"
	CreatedAt    time.Time `json:"created_at"`
}

// Routine represents an exercise routine assigned to a student
type Routine struct {
	ID         int64      `json:"id"`
	UserID    int64      `json:"user_id"`    // student who owns the routine
	CreatedBy int64      `json:"created_by"` // professor who created it
	Title     string     `json:"title"`
	StartDate string     `json:"start_date,omitempty"`
	EndDate   string     `json:"end_date,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	Exercises []Exercise `json:"exercises,omitempty"`
}

// Exercise represents a single exercise in a routine
type Exercise struct {
	ID            int64     `json:"id"`
	RoutineID    int64     `json:"routine_id"`
	DayNumber    int       `json:"day_number"`
	Name        string    `json:"name"`
	Sets        int       `json:"sets"`
	Reps        string   `json:"reps"` // can be "12" or "12 c/p"
	WeightKg    string   `json:"weight_kg,omitempty"` // can be "20 22 26" for progressive
	Observations string   `json:"observations,omitempty"`
	SortOrder   int      `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

// ExerciseLog represents a student's log for a specific exercise session
type ExerciseLog struct {
	ID           int64     `json:"id"`
	ExerciseID  int64     `json:"exercise_id"`
	UserID      int64     `json:"user_id"`
	Date        string    `json:"date"`
	Completed   bool     `json:"completed"`
	ActualWeight string   `json:"actual_weight,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Progress represents a student's progress over time for an exercise
type Progress struct {
	ExerciseID   int64    `json:"exercise_id"`
	ExerciseName string   `json:"exercise_name"`
	Logs         []ExerciseLog `json:"logs"`
}

// --- Request/Response DTOs ---

type RegisterRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	Name         string `json:"name" binding:"required"`
	Role         string `json:"role" binding:"required,oneof=student professor"`
	ProfessorCode string `json:"professor_code,omitempty"` // required if role is professor
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateRoutineRequest struct {
	UserID    int64  `json:"user_id" binding:"required"` // student
	Title     string `json:"title" binding:"required"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

type UpdateRoutineRequest struct {
	Title     string `json:"title,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
}

type CreateExerciseRequest struct {
	DayNumber    int     `json:"day_number" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Sets        int     `json:"sets" binding:"required,min=1"`
	Reps        string  `json:"reps" binding:"required"`
	WeightKg    string  `json:"weight_kg,omitempty"`
	Observations string `json:"observations,omitempty"`
	SortOrder   int     `json:"sort_order,omitempty"`
}

type UpdateExerciseRequest struct {
	DayNumber    *int     `json:"day_number,omitempty"`
	Name        *string   `json:"name,omitempty"`
	Sets        *int      `json:"sets,omitempty"`
	Reps        *string  `json:"reps,omitempty"`
	WeightKg    *string  `json:"weight_kg,omitempty"`
	Observations *string  `json:"observations,omitempty"`
	SortOrder   *int     `json:"sort_order,omitempty"`
}

type LogExerciseRequest struct {
	Date         string `json:"date" binding:"required"`
	Completed   bool   `json:"completed"`
	ActualWeight string `json:"actual_weight,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type APIResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Code  string `json:"code,omitempty"`
	Total int    `json:"total,omitempty"`
}