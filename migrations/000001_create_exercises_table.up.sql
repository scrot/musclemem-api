 CREATE TABLE IF NOT EXISTS exercises (
  exercise_id INTEGER PRIMARY KEY,
  owner INTEGER NOT NULL,
  workout INTEGER NOT NULL,
  name TEXT NOT NULL,
  weight REAL,
  repetitions INTEGER,
  next INTEGER,
  previous INTEGER,
  FOREIGN KEY (owner)
    REFERENCES users (user_id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,
  FOREIGN KEY (workout)
    REFERENCES workouts (workout_id)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);
