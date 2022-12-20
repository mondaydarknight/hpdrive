package infrastructure

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mondaydarknight/hpdrive/internal/domain/entity"
)

const createTableQuery = `CREATE TABLE IF NOT EXISTS files (
	id INTEGER NOT NULL PRIMARY KEY,
	dir VARCHAR(100) NOT NULL,
	fileName VARCHAR(100) NOT NULL,
	size UNSIGNED MEDIUMINT NOT NULL DEFAULT 0,
	content BLOB NOT NULL DEFAULT (x''),
	isArchived BOOLEAN NOT NULL DEFAULT 0 CHECK (isArchived IN (0, 1)),
	createdAt DATETIME NOT NULL,
	lastModified DATETIME NOT NULL
);
CREATE INDEX filePath ON files(dir, fileName);`

type FileRepository struct {
	db *sql.DB
}

// Create a new repository instance.
func NewFileRepository() (*FileRepository, error) {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(createTableQuery); err != nil {
		return nil, err
	}
	return &FileRepository{db}, nil
}

// Get the file entity by the directory and filename.
func (r *FileRepository) GetByFilePath(dir, fileName string) (*entity.File, error) {
	file := entity.File{}
	row := r.db.QueryRow("SELECT id, dir, fileName, size, content FROM files WHERE isArchived = ? AND dir = ? AND fileName = ?", false, dir, fileName)
	if err := row.Scan(&file.Id, &file.Dir, &file.FileName, &file.Size, &file.Content); err != nil {
		return nil, err
	}
	return &file, nil
}

// Get a list of files by the directories.
func (r *FileRepository) GetByDir(dir, ignore, orderBy, direction string) ([]*entity.File, error) {
	var files []*entity.File
	sql := "SELECT id, dir, fileName, size FROM files WHERE isArchived = ? AND dir = ?"
	if ignore != "" {
		ignore = "%" + ignore + "%"
		sql = fmt.Sprintf("%s AND fileName LIKE ?", sql)
	}
	if orderBy != "" {
		sql = fmt.Sprintf("%s ORDER BY %s %s", sql, orderBy, direction)
	}
	rows, err := r.db.Query(sql, false, dir, ignore)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		file := entity.File{}
		if err := rows.Scan(&file.Id, &file.Dir, &file.FileName, &file.Size); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return files, nil
}

// Create a file entity to the persistence.
func (r *FileRepository) Create(f *entity.File) (*entity.File, error) {
	now := time.Now()
	if idx := strings.LastIndex(f.Dir, "/"); idx > 0 {
		dir, subDir := f.Dir[:idx], f.Dir[idx+1:]
		if _, err := r.GetByFilePath(dir, subDir); err == sql.ErrNoRows {
			r.db.Exec("INSERT INTO files (dir, fileName, size, createdAt, lastModified) VALUES (?, ?, ?, ?, ?)", dir, subDir, 0, now, now)
		}
	}
	res, err := r.db.Exec("INSERT INTO files (dir, fileName, size, content, createdAt, lastModified) VALUES (?, ?, ?, ?, ?, ?)",
		f.Dir, f.FileName, f.Size, f.Content, now, now)
	if err != nil {
		return f, err
	}
	if f.Id, err = res.LastInsertId(); err != nil {
		return f, err
	}
	return f, nil
}

// Update the file content by ID.
func (r *FileRepository) Update(f *entity.File) error {
	stmt, err := r.db.Prepare("UPDATE files SET size = ?, content = ?, lastModified = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(f.Size, f.Content, time.Now(), f.Id)
	if err != nil {
		return err
	}
	return nil
}

// Mark the file as the archived.
func (r *FileRepository) MarkAsArchived(dir, fileName string) error {
	stmt, err := r.db.Prepare("UPDATE files set isArchived = ? WHERE dir = ? AND fileName = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(1, dir, fileName)
	if err != nil {
		return err
	}
	return nil
}
