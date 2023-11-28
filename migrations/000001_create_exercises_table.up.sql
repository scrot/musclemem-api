 CREATE TABLE IF NOT EXISTS exercises (
  exercise_id INTEGER PRIMARY KEY,
  workout INTEGER NOT NULL,
  name TEXT NOT NULL,
  weight REAL NOT NULL,
  repetitions INTEGER NOT NULL,
  next INTEGER NOT NULL,
  previous INTEGER NOT NULL,
  FOREIGN KEY (workout)
    REFERENCES workouts (workout_id)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);
