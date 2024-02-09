CREATE TABLE IF NOT EXISTS users (
  username TEXT NOT NULL, 
  email TEXT NOT NULL UNIQUE, 
  password TEXT NOT NULL,
  PRIMARY KEY (username)
);
