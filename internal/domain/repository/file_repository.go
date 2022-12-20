package repository

import "github.com/mondaydarknight/hpdrive/internal/domain/entity"

type FileRepository interface {
	// Get a list of files by the directories.
	GetByDir(dir, ignore, orderBy, direction string) ([]*entity.File, error)
	// Get the file entity by the directory and filename.
	GetByFilePath(dir string, fileName string) (*entity.File, error)
	// Create a file entity to the persistence.
	Create(f *entity.File) (*entity.File, error)
	// Update the file content by ID.
	Update(f *entity.File) error
	// Mark the file as the archived.
	MarkAsArchived(dir, fileName string) error
}
