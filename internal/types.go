package internal

type File struct {
	ID       int64  `db:"id"`
	FilePath string `db:"file_path"`
	MD5Hash  string `db:"hash"`
}
