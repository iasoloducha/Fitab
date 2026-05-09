package services

import (
	"database/sql"
	"fmt"
	"time"

	"fitlab-api/internal/handlers"
)

type Notifier struct {
	DB              *sql.DB
	sentReminders    map[int64]bool
	sentCompletions map[int64]bool
}

func NewNotifier(db *sql.DB) *Notifier {
	return &Notifier{
		DB:              db,
		sentReminders:   make(map[int64]bool),
		sentCompletions: make(map[int64]bool),
	}
}

func (n *Notifier) findRoutinesExpiringInDays(days int) ([]routineInfo, error) {
	query := `
		SELECT r.id, r.user_id, r.created_by, r.title, r.end_date,
		       u.email, u.name,
		       p.email, p.name
		FROM routines r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users p ON r.created_by = p.id
		WHERE date(r.end_date) = date('now', '+' || ? || ' days')
		  AND r.is_active = 1`
	rows, err := n.DB.Query(query, days)
	if err != nil {
		return nil, fmt.Errorf("query expiring routines: %v", err)
	}
	defer rows.Close()

	var routines []routineInfo
	for rows.Next() {
		var ri routineInfo
		var createdBy sql.NullInt64
		var profEmail, profName sql.NullString
		if err := rows.Scan(&ri.ID, &ri.UserID, &createdBy, &ri.Title, &ri.EndDate,
			&ri.StudentEmail, &ri.StudentName, &profEmail, &profName); err != nil {
			continue
		}
		if createdBy.Valid {
			ri.ProfessorEmail = profEmail.String
			ri.ProfessorName = profName.String
		}
		routines = append(routines, ri)
	}
	return routines, nil
}

func (n *Notifier) findRoutinesCompletedToday() ([]routineInfo, error) {
	query := `
		SELECT r.id, r.user_id, r.created_by, r.title, r.start_date, r.end_date,
		       u.email, u.name,
		       p.email, p.name
		FROM routines r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users p ON r.created_by = p.id
		WHERE date(r.end_date) = date('now')
		  AND r.is_active = 1`
	rows, err := n.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query completed routines: %v", err)
	}
	defer rows.Close()

	var routines []routineInfo
	for rows.Next() {
		var ri routineInfo
		var createdBy sql.NullInt64
		var profEmail, profName sql.NullString
		if err := rows.Scan(&ri.ID, &ri.UserID, &createdBy, &ri.Title, &ri.StartDate, &ri.EndDate,
			&ri.StudentEmail, &ri.StudentName, &profEmail, &profName); err != nil {
			continue
		}
		if createdBy.Valid {
			ri.ProfessorEmail = profEmail.String
			ri.ProfessorName = profName.String
		}
		routines = append(routines, ri)
	}
	return routines, nil
}

func (n *Notifier) getProgressStats(userID int64, startDate, endDate string) (handlers.ProgressStats, error) {
	query := `
		SELECT COUNT(*) FROM exercise_logs el
		JOIN exercises e ON el.exercise_id = e.id
		JOIN routines r ON e.routine_id = r.id
		WHERE el.user_id = ? AND el.date BETWEEN ? AND ?`
	var total, completed int
	n.DB.QueryRow(query, userID, startDate, endDate).Scan(&total)
	n.DB.QueryRow(query+" AND el.completed = 1", userID, startDate, endDate).Scan(&completed)

	var rate float64
	if total > 0 {
		rate = float64(completed) / float64(total) * 100
	}

	weightsQuery := `
		SELECT e.name, el.actual_weight, el.date
		FROM exercise_logs el
		JOIN exercises e ON el.exercise_id = e.id
		WHERE el.user_id = ? AND el.date BETWEEN ? AND ?
		ORDER BY e.name, el.date`
	rows, err := n.DB.Query(weightsQuery, userID, startDate, endDate)
	if err != nil {
		return handlers.ProgressStats{}, err
	}
	defer rows.Close()

	var summary string
	var lastName string
	var weights map[string]string = make(map[string]string)
	for rows.Next() {
		var name, weight, date string
		rows.Scan(&name, &weight, &date)
		if name != lastName && lastName != "" {
			summary += fmt.Sprintf("- %s: %s\n", lastName, weights[lastName])
		}
		lastName = name
		weights[name] = weight
	}
	if lastName != "" {
		summary += fmt.Sprintf("- %s: %s\n", lastName, weights[lastName])
	}

	return handlers.ProgressStats{
		ExercisesLogged: total,
		CompletionRate:  rate,
		PerExerciseData: summary,
	}, nil
}

func (n *Notifier) RunCheck() (int, int, error) {
	expiring, err := n.findRoutinesExpiringInDays(3)
	if err != nil {
		return 0, 0, err
	}

	sentReminders := 0
	for _, ri := range expiring {
		if n.sentReminders[ri.ID] {
			continue
		}
		err := handlers.SendExpirationReminderEmail(
			ri.StudentEmail, ri.StudentName,
			ri.ProfessorEmail, ri.Title, ri.EndDate,
		)
		if err != nil {
			fmt.Printf("ERROR sending expiration reminder for routine %d: %v\n", ri.ID, err)
			continue
		}
		n.sentReminders[ri.ID] = true
		sentReminders++
	}

	completed, err := n.findRoutinesCompletedToday()
	if err != nil {
		return sentReminders, 0, err
	}

	sentCompletions := 0
	for _, ri := range completed {
		if n.sentCompletions[ri.ID] {
			continue
		}
		stats, err := n.getProgressStats(ri.UserID, ri.StartDate, ri.EndDate)
		if err != nil {
			fmt.Printf("ERROR getting progress stats for routine %d: %v\n", ri.ID, err)
			continue
		}
		err = handlers.SendRoutineCompletedEmail(
			ri.StudentEmail, ri.StudentName,
			ri.ProfessorName, ri.Title, stats,
		)
		if err != nil {
			fmt.Printf("ERROR sending completion email for routine %d: %v\n", ri.ID, err)
			continue
		}
		n.sentCompletions[ri.ID] = true
		sentCompletions++
	}

	return sentReminders, sentCompletions, nil
}

type routineInfo struct {
	ID              int64
	UserID          int64
	Title           string
	StartDate       string
	EndDate         string
	StudentEmail    string
	StudentName     string
	ProfessorEmail  string
	ProfessorName   string
}

func StartNotifierLoop(db *sql.DB, interval time.Duration) {
	notifier := NewNotifier(db)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Printf("[scheduler] Notifier started, checking every %v\n", interval)

	for {
		now := time.Now().Format("2006-01-02 15:04:05")
		reminders, completions, err := notifier.RunCheck()
		if err != nil {
			fmt.Printf("[%s] ERROR: %v\n", now, err)
		} else if reminders > 0 || completions > 0 {
			fmt.Printf("[%s] Sent %d reminders, %d completions\n", now, reminders, completions)
		} else {
			fmt.Printf("[%s] No emails to send\n", now)
		}

		<-ticker.C
	}
}
