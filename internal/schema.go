package internal

var Schema = `PRAGMA journal_mode = WAL;
	CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    md5_hash CHAR(32) NOT NULL
)
`
