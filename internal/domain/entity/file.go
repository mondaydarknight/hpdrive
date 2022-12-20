package entity

import "time"

type File struct {
	Id        int64
	Dir       string
	FileName  string
	Size      int
	Content   []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}
