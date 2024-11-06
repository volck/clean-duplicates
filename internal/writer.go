package internal

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	sqlx "github.com/jmoiron/sqlx"
)

type Writer struct {
	WriteChan chan File
}

func NewWriter(writerChan chan File) *Writer {
	return &Writer{WriteChan: writerChan}

}

func (w *Writer) Listen(inChan <-chan File, wg *sync.WaitGroup) {
	defer wg.Done()

	db, err := w.OpenDb()
	if err != nil {
		Logger.Error("could not open database", slog.Any("error", err))
	}
	Logger.Info("writer listening")
	counter := 0
	for file := range inChan {
		counter++
		Logger.Info("writer recieved file", slog.Any("file", file), slog.Int("counter", counter))
		exists, err := w.FileAlreadyChecked(db, file.FilePath)
		if err != nil {
			Logger.Error("could not check if file exists", slog.Any("error", err))
		}
		if !exists {
			w.InsertFile(db, &file)
		} else {
			Logger.Info("file already exists", slog.Any("file", file))
		}
	}
	Logger.Info("writer done")
}

func (w *Writer) InitDB() {

	db := w.makeDb()
	Logger.Info("db stats", slog.Int("inuse", db.Stats().InUse), slog.Any("ping", db.Ping()))

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv

	_, err := db.Exec(Schema)
	if err != nil {
		Logger.Error("failed to exec", slog.Any("error", err))
	}
	res, err := db.Exec("PRAGMA table_info(files)")
	if err != nil {
		Logger.Error("could not describe files", slog.Any("error", err))
	}
	Logger.Info("describe files", slog.Any("res", res))
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

func (w *Writer) InsertFile(db *sqlx.DB, f *File) error {
	res, err := db.Exec("INSERT INTO files (file_path, hash) VALUES (?, ?)", f.FilePath, f.MD5Hash)
	if err != nil {
		Logger.Error("could not insert file", slog.Any("error", err))
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		Logger.Error("could not get rows affected", slog.Any("error", err))
	}

	Logger.Info("inserted file", slog.Any("rowsAffected", rowsAffected))
	return err
}

func (w *Writer) FileAlreadyChecked(db *sqlx.DB, path string) (bool, error) {

	query := "SELECT 1 FROM files WHERE file_path = ? LIMIT 1"

	var exists bool
	err := db.QueryRow(query, path).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return exists, err
}

func (w *Writer) GetDuplicates() []File {

	db, err := OpenDb()
	query := `SELECT f1.id, f1.filename, f1.file_path, f1.md5_hash
FROM files f1
INNER JOIN (
    SELECT md5_hash
    FROM files
    GROUP BY md5_hash
    HAVING COUNT(*) > 1
) f2 ON f1.md5_hash = f2.md5_hash
ORDER BY f1.md5_hash, f1.filename
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
