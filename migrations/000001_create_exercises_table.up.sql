 CREATE TABLE IF NOT EXISTS exercises (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  weight REAL,
  repetitions INTEGER,
  next TEXT,
  previous TEXT
);
