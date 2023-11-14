 CREATE TABLE IF NOT EXISTS exercises (
  excercise_id INTEGER PRIMARY KEY,
  owner INTEGER NOT NULL,
  workout TEXT NOT NULL,
  name TEXT NOT NULL,
  weight REAL,
  repetitions INTEGER,
  next TEXT,
  previous TEXT
);
