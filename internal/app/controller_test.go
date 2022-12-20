package app

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/mondaydarknight/hpdrive/internal/domain/entity"
)

func TestGetFile(t *testing.T) {
	tests := []struct {
		path        string
		file        *entity.File
		expectedErr error
	}{
		{"/file/foo?orderBy=foo", nil, errors.New("[orderBy] field foo is not allowed")},
		{"/file/foo?orderDirection=foo", nil, errors.New("[orderDirection] field foo is not allowed")},
		{"/file/foo?orderBy=fileName", nil, errors.New("Both [orderBy] and [orderDirection] fields must be specified")},
		{"/file/foo", nil, errors.New("no files matching file path")},
		{"/file/foo", &entity.File{FileName: "baz.txt"}, nil},
		{"/file/foo.bar.txt", nil, sql.ErrNoRows},
		{"/file/foo.bar.txt", &entity.File{Content: []byte{0, 1, 2}}, nil},
	}
	for _, tt := range tests {
		r, err := http.NewRequest("GET", tt.path, bytes.NewBuffer(nil))
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		c := &controller{&mockFileRepoistory{tt.file}}
		err = c.get(w, r)
		if !errors.Is(err, tt.expectedErr) {
			t.Errorf("expected error (%v), got error (%v)", tt.expectedErr, err)
		}
	}
}

func TestCreateFile(t *testing.T) {
	tests := []struct {
		path        string
		file        []byte
		existedFile *entity.File
		expectedErr error
	}{
		{"/file/foo", nil, nil, errors.New("file /file/foo must be a file")},
		{"/file/foo/bar.txt", []byte{0, 1, 2}, &entity.File{}, errors.New("file /file/foo/bar.txt has already existed")},
		{"/file/foo/bar.txt", []byte{0, 1, 2}, nil, nil},
	}
	for _, tt := range tests {
		f := bytes.NewBuffer(tt.file)
		b := &bytes.Buffer{}
		m := multipart.NewWriter(b)
		part, _ := m.CreateFormFile("file", filepath.Base(tt.path))
		io.Copy(part, f)
		m.Close()
		r, err := http.NewRequest("POST", tt.path, b)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", m.FormDataContentType())
		w := httptest.NewRecorder()
		c := &controller{&mockFileRepoistory{tt.existedFile}}
		err = c.create(w, r)
		if !errors.Is(err, tt.expectedErr) {
			t.Errorf("expected error (%v), got error (%v)", tt.expectedErr, err)
		}
	}
}

func TestUpdateFile(t *testing.T) {
	tests := []struct {
		path        string
		file        []byte
		existedFile *entity.File
		expectedErr error
	}{
		{"/file/foo", nil, nil, errors.New("file /file/foo must be a file")},
		{"/file/foo/bar.txt", []byte{0, 1, 2}, nil, sql.ErrNoRows},
		{"/file/foo/bar.txt", []byte{0, 1, 2}, &entity.File{}, nil},
	}
	for _, tt := range tests {
		f := bytes.NewBuffer(tt.file)
		b := &bytes.Buffer{}
		m := multipart.NewWriter(b)
		part, _ := m.CreateFormFile("file", filepath.Base(tt.path))
		io.Copy(part, f)
		m.Close()
		r, err := http.NewRequest("PATCH", tt.path, b)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", m.FormDataContentType())
		w := httptest.NewRecorder()
		c := &controller{&mockFileRepoistory{tt.existedFile}}
		err = c.patch(w, r)
		if !errors.Is(err, tt.expectedErr) {
			t.Errorf("expected error (%v), got error (%v)", tt.expectedErr, err)
		}
	}
}

func TestDeleteFile(t *testing.T) {
	tests := []struct {
		path        string
		expectedErr error
	}{
		{"/file/foo", errors.New("file /file/foo must be a file")},
		{"/file/foo/bar.txt", nil},
	}
	for _, tt := range tests {
		r, err := http.NewRequest("DELETE", tt.path, bytes.NewBuffer(nil))
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		c := &controller{&mockFileRepoistory{}}
		err = c.delete(w, r)
		if !errors.Is(err, tt.expectedErr) {
			t.Errorf("expected error (%v), got error (%v)", tt.expectedErr, err)
		}
	}
}

type mockFileRepoistory struct {
	file *entity.File
}

func (r *mockFileRepoistory) GetByDir(dir, ignore, orderBy, direction string) ([]*entity.File, error) {
	if r.file == nil {
		return nil, nil
	}
	return []*entity.File{r.file}, nil
}

func (r *mockFileRepoistory) GetByFilePath(dir string, fileName string) (*entity.File, error) {
	if r.file == nil {
		return nil, sql.ErrNoRows
	}
	return r.file, nil
}

func (r *mockFileRepoistory) Create(f *entity.File) (*entity.File, error) {
	return nil, nil
}

func (r *mockFileRepoistory) Update(f *entity.File) error {
	return nil
}

func (r *mockFileRepoistory) MarkAsArchived(dir, fileName string) error {
	return nil
}
