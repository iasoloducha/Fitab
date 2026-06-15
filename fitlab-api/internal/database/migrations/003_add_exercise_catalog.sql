-- Exercise Catalog Migration
-- Creates a global exercise catalog and links routine exercises to it

CREATE TABLE IF NOT EXISTS exercise_catalog (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    image_urls TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE exercises ADD COLUMN catalog_exercise_id INTEGER REFERENCES exercise_catalog(id);
CREATE INDEX IF NOT EXISTS idx_exercises_catalog ON exercises(catalog_exercise_id);

-- Seed catalog with default exercises from routine.go
INSERT INTO exercise_catalog (name) VALUES
('Press de banca plano'),
('Press de banca inclinado'),
('Aperturas con mancuernas'),
('Fondos en paralelas'),
('Extensiones de triceps polea'),
('Extensiones de triceps cabeza'),
('Dominadas o Jalón al pecho'),
('Remo con barra'),
('Remo con mancuernas'),
('Curl con barra'),
('Curl con mancuernas alternado'),
('Curl martillo'),
('Sentadillas'),
('Prensa de piernas'),
('Extensiones de cuádriceps'),
('Curl de femoral'),
('Elevación de talones'),
('Abducción de cadera');

-- Backfill existing exercises that match catalog names
UPDATE exercises
SET catalog_exercise_id = (
    SELECT ec.id FROM exercise_catalog ec
    WHERE ec.name = exercises.name
)
WHERE exercises.name IN (SELECT name FROM exercise_catalog);
