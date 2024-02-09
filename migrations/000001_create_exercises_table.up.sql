 CREATE TABLE IF NOT EXISTS exercises (
  owner TEXT NOT NULL,
  workout INTEGER NOT NULL,
  exercise_index INTEGER NOT NULL, 
  name TEXT NOT NULL,
  weight REAL NOT NULL,
  repetitions INTEGER NOT NULL,
  PRIMARY KEY (owner, workout, exercise_index),
  FOREIGN KEY (owner, workout)
    REFERENCES workouts (owner, workout_index)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);
