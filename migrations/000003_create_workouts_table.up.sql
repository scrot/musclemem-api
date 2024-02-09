CREATE TABLE IF NOT EXISTS workouts (
  owner TEXT NOT NULL,
  workout_index INTEGER NOT NULL,
  name TEXT NOT NULL,
  PRIMARY KEY (owner, workout_index),
  FOREIGN KEY (owner)
    REFERENCES users (username)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);
