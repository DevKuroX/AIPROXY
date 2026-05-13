package v1

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/google/uuid"
)

// ref: OpenAI Files API

type FileObject struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int64  `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

type FileStore interface {
	CreateFile(ctx context.Context, filename, purpose string, data []byte) (*FileObject, error)
	GetFile(ctx context.Context, fileID string) (*FileObject, error)
	GetFileContent(ctx context.Context, fileID string) ([]byte, error)
	ListFiles(ctx context.Context, purpose string) ([]FileObject, error)
	DeleteFile(ctx context.Context, fileID string) error
}

var fileStore FileStore

func SetFileStore(store FileStore) {
	fileStore = store
}

func HandleListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	purpose := r.URL.Query().Get("purpose")

	if fileStore == nil {
		errs.WriteJSONError(w, "file storage not configured", http.StatusServiceUnavailable)
		return
	}

	files, err := fileStore.ListFiles(ctx, purpose)
	if err != nil {
		errs.WriteJSONError(w, "failed to list files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   files,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if fileStore == nil {
		errs.WriteJSONError(w, "file storage not configured", http.StatusServiceUnavailable)
		return
	}

	err := r.ParseMultipartForm(512 << 20)
	if err != nil {
		errs.WriteJSONError(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		errs.WriteJSONError(w, "missing required field: file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	purpose := r.FormValue("purpose")
	if purpose == "" {
		errs.WriteJSONError(w, "missing required field: purpose", http.StatusBadRequest)
		return
	}

	validPurposes := map[string]bool{
		"assistants":        true,
		"assistants_output": true,
		"batch":             true,
		"fine-tune":         true,
		"vision":            true,
	}
	if !validPurposes[purpose] {
		errs.WriteJSONError(w, "invalid purpose. Must be one of: assistants, assistants_output, batch, fine-tune, vision", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		errs.WriteJSONError(w, "failed to read file content", http.StatusInternalServerError)
		return
	}

	fileObj, err := fileStore.CreateFile(ctx, header.Filename, purpose, data)
	if err != nil {
		errs.WriteJSONError(w, "failed to store file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fileObj)
}

func HandleGetFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("file_id")

	if fileID == "" {
		errs.WriteJSONError(w, "file_id is required", http.StatusBadRequest)
		return
	}

	if fileStore == nil {
		errs.WriteJSONError(w, "file storage not configured", http.StatusServiceUnavailable)
		return
	}

	fileObj, err := fileStore.GetFile(ctx, fileID)
	if err != nil {
		errs.WriteJSONError(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileObj)
}

func HandleGetFileContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("file_id")

	if fileID == "" {
		errs.WriteJSONError(w, "file_id is required", http.StatusBadRequest)
		return
	}

	if fileStore == nil {
		errs.WriteJSONError(w, "file storage not configured", http.StatusServiceUnavailable)
		return
	}

	content, err := fileStore.GetFileContent(ctx, fileID)
	if err != nil {
		errs.WriteJSONError(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(content)
}

func HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("file_id")

	if fileID == "" {
		errs.WriteJSONError(w, "file_id is required", http.StatusBadRequest)
		return
	}

	if fileStore == nil {
		errs.WriteJSONError(w, "file storage not configured", http.StatusServiceUnavailable)
		return
	}

	err := fileStore.DeleteFile(ctx, fileID)
	if err != nil {
		errs.WriteJSONError(w, "failed to delete file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":      fileID,
		"object":  "file",
		"deleted": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type InMemoryFileStore struct {
	files map[string]*storedFile
}

type storedFile struct {
	FileObject
	content []byte
}

func NewInMemoryFileStore() *InMemoryFileStore {
	return &InMemoryFileStore{
		files: make(map[string]*storedFile),
	}
}

func (s *InMemoryFileStore) CreateFile(ctx context.Context, filename, purpose string, data []byte) (*FileObject, error) {
	id := "file-" + uuid.New().String()[:24]

	fileObj := &FileObject{
		ID:        id,
		Object:    "file",
		Bytes:     int64(len(data)),
		CreatedAt: time.Now().Unix(),
		Filename:  filename,
		Purpose:   purpose,
	}

	s.files[id] = &storedFile{
		FileObject: *fileObj,
		content:    data,
	}

	return fileObj, nil
}

func (s *InMemoryFileStore) GetFile(ctx context.Context, fileID string) (*FileObject, error) {
	if f, ok := s.files[fileID]; ok {
		return &f.FileObject, nil
	}
	return nil, errors.New("file not found")
}

func (s *InMemoryFileStore) GetFileContent(ctx context.Context, fileID string) ([]byte, error) {
	if f, ok := s.files[fileID]; ok {
		return f.content, nil
	}
	return nil, errors.New("file not found")
}

func (s *InMemoryFileStore) ListFiles(ctx context.Context, purpose string) ([]FileObject, error) {
	var result []FileObject
	for _, f := range s.files {
		if purpose == "" || f.Purpose == purpose {
			result = append(result, f.FileObject)
		}
	}
	return result, nil
}

func (s *InMemoryFileStore) DeleteFile(ctx context.Context, fileID string) error {
	delete(s.files, fileID)
	return nil
}
