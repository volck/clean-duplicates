package internal

type File struct {
	ID       int64  `db:"id"`
	FilePath string `db:"file_path"`
	Hash     string `db:"hash"`
}
