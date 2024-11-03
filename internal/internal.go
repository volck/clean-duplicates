package internal

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var PathsFound []string

func makeDB() (db *sqlx.DB) {
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

func InitDB() {

	db := makeDB()

	Logger.Info("db stats", slog.Int("inuse", db.Stats().InUse), slog.Any("ping", db.Ping()))

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	_, err := db.Exec(Schema)
	if err != nil {
		Logger.Error("failed to exec", slog.Any("error", err))
	}

}

func FindFiles(directory string) {
	err := filepath.WalkDir(directory, walkFunc)
	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", directory, err)
	}

}

func countFiles(path string, info os.DirEntry, err error) error {
	if err != nil {
		fmt.Printf("Error accessing path %q: %v\n", path, err)
		return err
	}

	if info.IsDir() {
	} else {
		NumberOfFiles++
		return nil
	}
	return nil
}
func FindNumberOfFiles(directory string) {
	err := filepath.WalkDir(directory, countFiles)
	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", directory, err)
	}

}

func CalculateFile(path string, db *sqlx.DB) File {
	theHash, err := calculateHash(path)
	if err != nil {
		Logger.Error("could not calculate hashes", slog.Any("error", err))
	}
	md5Hash := fmt.Sprintf("%x", *theHash)
	f := File{FilePath: path, MD5Hash: md5Hash}
	Logger.Info("calculated hash", slog.Any("hash", md5Hash))
	err = InsertFile(db, &f)
	if err != nil {
		Logger.Error("could not insert file", slog.Any("error", err))
	}
	Logger.Info("done calculating file", slog.Any("file", f))
	Wg.Done()
	return f
}

func walkFunc(path string, info os.DirEntry, err error) error {

	if err != nil {
		fmt.Printf("Error accessing path %q: %v\n", path, err)
		return err
	}

	if info.IsDir() {
	} else {
		PathsFound = append(PathsFound, path)
	}
	return nil
}

func calculateHash(filePath string) (md5sum *[]byte, err error) {
	Logger.Info("calculating hash", slog.Any("path", filePath))
	file, err := os.Open(filePath)
	if err != nil {
		Logger.Error("error opening file", slog.Any("error", err))
		return
	}

	sha256Hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(sha256Hash), file); err != nil {
		Logger.Error("io copy on calculateHash failed", slog.Any("err", err))
		return nil, err
	}

	theHash := sha256Hash.Sum(nil)

	err = file.Close()
	if err != nil {
		Logger.Error("Could not close file", slog.Any("error", err))

	}

	Logger.Info("calculated hash", slog.Any("hash", theHash))
	return &theHash, nil

}
func OpenDb() (*sqlx.DB, error) {
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

func InsertFile(db *sqlx.DB, file *File) error {
	// Start a transaction
	Logger.Info("inserting file", slog.Any("file", file))
	tx, err := db.Beginx()
	if err != nil {
		Logger.Error("failed to begin transaction", slog.Any("error", err))
		return err
	}

	// Defer a rollback in case anything fails
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				Logger.Error("failed to rollback transaction", slog.Any("error", rbErr))
			}
		}
	}()

	// Execute the insert within the transaction
	result, err := tx.NamedExec(`
        INSERT INTO files (file_path, md5_hash)
        VALUES (:file_path, :md5_hash)
    `, file)
	if err != nil {
		Logger.Error("failed to insert file", slog.Any("error", err))
		return err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		Logger.Error("failed to get last insert ID", slog.Any("error", err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		Logger.Error("failed to get rows affected", slog.Any("error", err))
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		Logger.Error("failed to commit transaction", slog.Any("error", err))
		return err
	}

	Logger.Info("inserted file",
		slog.Any("lastId", lastId),
		slog.Any("rowsAffected", rowsAffected))

	return nil
}

func FileAlreadyChecked(path string) (bool, error) {

	db, err := OpenDb()
	if err != nil {
		Logger.Error("error", slog.Any("error", err))
	}
	query := "SELECT 1 FROM files WHERE file_path = ? LIMIT 1"

	var exists bool
	err = db.QueryRow(query, path).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil // Filename doesn't exist

	} else if err != nil {
		return false, err // An error occurred
	}

	Logger.Info("File already exists", slog.Any("path", path))
	return true, nil // Filename exists
}

func GetDuplicates() []File {

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
