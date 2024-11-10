package internal

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func FindFiles(directory string) {
	err := filepath.WalkDir(directory, walkFunc)
	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", directory, err)
	}

}

func walkFunc(path string, info os.DirEntry, err error) error {

	if err != nil {
		fmt.Printf("Error accessing path %q: %v\n", path, err)
		return err
	}

	if info.IsDir() {
	} else {

	}
	return nil
}

func calculateHash(filePath string) (md5sum *[]byte, err error) {
	Logger.Debug("Calculating hash for file", slog.String("file", filePath))
	file, err := os.Open(filePath)
	if err != nil {
		Logger.Error("error opening file", slog.Any("error", err))
		return
	}

	defer file.Close()
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
