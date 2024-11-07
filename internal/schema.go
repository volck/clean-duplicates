package internal

var Schema = `PRAGMA journal_mode = WAL;
	CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT UNIQUE NOT NULL,
    hash CHAR(32) NOT NULL
)
`
