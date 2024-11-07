package internal

import (
	"database/sql"
	"fmt"
	sqlx "github.com/jmoiron/sqlx"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	flushCount  = 500
	flushPeriod = 1 * time.Second
)

type Writer struct {
	WriteChan chan File
}

func NewWriter(writerChan chan File) *Writer {
	return &Writer{WriteChan: writerChan}

}

func (w *Writer) InitDB() {

	db := w.makeDb()
	Logger.Info("db stats", slog.Int("inuse", db.Stats().InUse), slog.Any("ping", db.Ping()))

	_, err := db.Exec(Schema)
	if err != nil {
		Logger.Error("failed to exec", slog.Any("error", err))
	}
}

func (w *Writer) makeDb() (db *sqlx.DB) {

	usercfgdir, err := os.UserConfigDir()
	if err != nil {
		Logger.Error("could not get user cfg dir")
	}

	os.Mkdir(fmt.Sprintf("%s/%s", usercfgdir, "clean-duplicate"), 0700)
	dbName := filepath.Join(usercfgdir, "clean-duplicate", "clean-duplicate.db")

	db, err = sqlx.Connect("sqlite3", dbName)
	if err != nil {
		Logger.Error("error", slog.Any("error", err))
	}
	return db

}

func (w *Writer) OpenDb() (*sqlx.DB, error) {
	usercfgdir, err := os.UserConfigDir()
	if err != nil {
		Logger.Error("could not get user cfg dir")
	}
	dbName := filepath.Join(usercfgdir, "clean-duplicate", "clean-duplicate.db")
	db, err := sqlx.Open("sqlite3", dbName)
	if err != nil {
		Logger.Error("could not open database", slog.Any("error", err))
	}
	return db, err

}
func (w *Writer) Listen(inChan <-chan File, wg *sync.WaitGroup) {
	defer wg.Done()

	db, err := w.OpenDb()
	if err != nil {
		Logger.Error("could not open database", slog.Any("error", err))
		return
	}

	Logger.Info("writer listening")

	tx, err := db.Begin()
	if err != nil {
		Logger.Error("could not begin transaction", slog.Any("error", err))
		return
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO files (file_path, hash) VALUES (?,?)")
	if err != nil {
		Logger.Error("could not prepare statement", slog.Any("error", err))
		return
	}
	defer stmt.Close()

	var files []File
	ticker := time.NewTicker(flushPeriod)
	defer ticker.Stop()

	for {
		select {
		case file, ok := <-inChan:
			if !ok {
				flushFiles(stmt, files)
				err = tx.Commit()
				if err != nil {
					Logger.Error("commit failed", slog.Any("error", err))
				}
				Logger.Info("writer done")
				return
			}
			files = append(files, file)
			// Flush hvis bufferstørrelsen er nådd
			if len(files) >= flushCount {
				Logger.Info("flushing files", slog.Int("count", len(files)))
				flushFiles(stmt, files)
				files = nil
			}
		case <-ticker.C:
			// Tidsbasert flush
			if len(files) > 0 {
				flushFiles(stmt, files)
				files = nil
			}
		}
	}
}
func flushFiles(stmt *sql.Stmt, files []File) {
	for _, file := range files {
		res, err := stmt.Exec(file.FilePath, file.Hash)
		if err != nil {
			Logger.Error("could not execute statement", slog.Any("error", err))
		}
		RowsAffected, err := res.RowsAffected()
		if err != nil {

			Logger.Error("could not get rows affected", slog.Any("error", err))
		}
		if RowsAffected > 0 {
			Logger.Info("inserted", slog.Any("file_path", file.FilePath), slog.Any("RowsAffected", RowsAffected))
		}
	}
}

func (w *Writer) GetDuplicates() []File {

	db, err := OpenDb()
	if err != nil {
		Logger.Error("could not open database", slog.Any("error", err))
	}
	query := `SELECT f1.id, f1.file_path, f1.hash
FROM files f1
INNER JOIN (
    SELECT hash
    FROM files
    GROUP BY hash
    HAVING COUNT(*) > 1
) f2 ON f1.hash = f2.hash
ORDER BY f1.hash, f1.file_path;
`

	var files []File
	err = db.Select(&files, query)
	if err != nil {
		Logger.Error("could not get duplicates", slog.Any("error", err))
	}

	return files

}

func (w *Writer) DeleteDb() {

	usercfgdir, err := os.UserConfigDir()
	if err != nil {
		Logger.Error("could not get user cfg dir")
	}

	dbName := filepath.Join(usercfgdir, "clean-duplicate", "clean-duplicate.db")
	err = os.Remove(dbName)
	if err != nil {
		Logger.Error("could not delete db", slog.Any("error", err))
	}
	Logger.Info("deleted db")
}
