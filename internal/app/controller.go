package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/mondaydarknight/hpdrive/internal/domain/entity"
	"github.com/mondaydarknight/hpdrive/internal/domain/repository"
	"golang.org/x/exp/slices"
)

const (
	filePrefix    = "/file"
	maxUploadSize = 10 << 20
)

type controller struct {
	repo repository.FileRepository
}

// Get file binary content or a list of files and sub-directories.
func (c *controller) get(w http.ResponseWriter, r *http.Request) error {
	dir, fileName := strings.TrimPrefix(path.Dir(r.URL.Path), filePrefix), path.Base(r.URL.Path)
	ob, od := r.URL.Query().Get("orderBy"), r.URL.Query().Get("orderDirection")
	if ob != "" && !slices.Contains([]string{"fileName", "size", "lastModified"}, ob) {
		return &appError{http.StatusBadRequest, fmt.Sprintf("[orderBy] field %s is not allowed", ob)}
	}
	if od != "" && !slices.Contains([]string{"Ascending", "Descending"}, od) {
		return &appError{http.StatusBadRequest, fmt.Sprintf("[orderDirection] field %s is not allowed", od)}
	}
	if (ob == "" && od != "") || (ob != "" && od == "") {
		return &appError{http.StatusBadRequest, "Both [orderBy] and [orderDirection] fields must be specified"}
	}
	if filepath.Ext(fileName) == "" {
		od = OrderDirectionMap[od].String()
		files, err := c.repo.GetByDir(dir+"/"+fileName, r.URL.Query().Get("filterByName"), ob, od)
		if err != nil {
			return &appError{http.StatusInternalServerError, err.Error()}
		}
		if files == nil {
			return &appError{http.StatusNotFound, "no files matching file path"}
		}
		resp := Dir{IsDirectory: true}
		for _, file := range files {
			resp.Files = append(resp.Files, file.FileName)
		}
		return replyJSON(w, resp, http.StatusOK)
	}
	file, err := c.repo.GetByFilePath(dir, fileName)
	if err != nil {
		return &appError{http.StatusNotFound, err.Error()}
	}
	w.Write(file.Content)
	return nil
}

// Create a uploaded file to the storage.
func (c *controller) create(w http.ResponseWriter, r *http.Request) error {
	dir, fileName := strings.TrimPrefix(path.Dir(r.URL.Path), filePrefix), path.Base(r.URL.Path)
	if filepath.Ext(fileName) == "" {
		return &appError{http.StatusBadRequest, fmt.Sprintf("file %s must be a file", r.URL.Path)}
	}
	r.ParseMultipartForm(maxUploadSize)
	f, _, err := r.FormFile("file")
	if err != nil {
		return &appError{http.StatusInternalServerError, err.Error()}
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		return &appError{http.StatusInternalServerError, err.Error()}
	}
	if _, err := c.repo.GetByFilePath(dir, fileName); err != sql.ErrNoRows {
		return &appError{http.StatusBadRequest, fmt.Sprintf("file %s has already existed", r.URL.Path)}
	}
	_, err = c.repo.Create(&entity.File{Dir: dir, FileName: fileName, Size: buf.Len(), Content: buf.Bytes()})
	if err != nil {
		return &appError{http.StatusInternalServerError, err.Error()}
	}
	return replyJSON(w, struct{}{}, http.StatusOK)
}

// Modify the existing file for the local storage.
func (c *controller) patch(w http.ResponseWriter, r *http.Request) error {
	dir, fileName := strings.TrimPrefix(path.Dir(r.URL.Path), filePrefix), path.Base(r.URL.Path)
	if filepath.Ext(fileName) == "" {
		return &appError{http.StatusBadRequest, fmt.Sprintf("file %s must be a file", r.URL.Path)}
	}
	r.ParseMultipartForm(maxUploadSize)
	f, _, err := r.FormFile("file")
	if err != nil {
		return &appError{http.StatusInternalServerError, err.Error()}
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		return &appError{http.StatusInternalServerError, err.Error()}
	}
	var file *entity.File
	if file, err = c.repo.GetByFilePath(dir, fileName); err == sql.ErrNoRows {
		return &appError{http.StatusNotFound, err.Error()}
	}
	file.Size, file.Content = buf.Len(), buf.Bytes()
	if err := c.repo.Update(file); err != nil {
		return &appError{http.StatusInternalServerError, err.Error()}
	}
	return replyJSON(w, struct{}{}, http.StatusOK)
}

// Delete an existing file from the local storage.
func (c *controller) delete(w http.ResponseWriter, r *http.Request) error {
	dir, fileName := strings.TrimPrefix(path.Dir(r.URL.Path), filePrefix), path.Base(r.URL.Path)
	if filepath.Ext(fileName) == "" {
		return &appError{http.StatusBadRequest, fmt.Sprintf("file %s must be a file", r.URL.Path)}
	}
	if err := c.repo.MarkAsArchived(dir, fileName); err != nil {
		return &appError{http.StatusInternalServerError, err.Error()}
	}
	return replyJSON(w, struct{}{}, http.StatusOK)
}

// Respond the output with JSON format to the client.
func replyJSON(w http.ResponseWriter, data interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	return nil
}
