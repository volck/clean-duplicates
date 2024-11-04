package internal

import (
	"fmt"
	sqlx "github.com/jmoiron/sqlx"
	"log/slog"
	"os"
	"path/filepath"
)

type Writer struct {
	WriteChan chan File
}

func NewWriter(writerChan chan File) *Writer {
	return &Writer{WriteChan: writerChan}

}

func (w *Writer) Listen() {
	Logger.Info("writer listening")
	for {
		select {
		case f := <-w.WriteChan:
			Logger.Info("received file", slog.Any("file", f))
			//w.InsertFile(f)
		}
	}
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
	_, err := db.Exec("INSERT INTO files (path, hash) VALUES (?, ?)", f.FilePath, f.MD5Hash)
	if err != nil {
		Logger.Error("could not insert file", slog.Any("error", err))
	}
	return err
}

func (w *Writer) FileAlreadyChecked(path string) (bool, error) {

	db, err := w.OpenDb()
	if err != nil {
		Logger.Error("error", slog.Any("error", err))
	}
	query := "SELECT 1 FROM files WHERE path = ? LIMIT 1"

	var exists bool
	err = db.QueryRow(query, path).Scan(&exists)
	if err != nil {
		Logger.Error("could not query db", slog.Any("error", err))
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

func (w *Writer) Writer() {

}
