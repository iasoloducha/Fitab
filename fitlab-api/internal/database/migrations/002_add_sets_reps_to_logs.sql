-- Fitlab API Migration 002
-- Add actual_sets and actual_reps columns to exercise_logs

ALTER TABLE exercise_logs ADD COLUMN actual_sets INTEGER;
ALTER TABLE exercise_logs ADD COLUMN actual_reps TEXT;
