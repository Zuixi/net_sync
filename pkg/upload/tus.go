package upload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/easy-sync/easy-sync/pkg/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tus/tusd/v2/pkg/handler"
)

type TusHandler struct {
	config   *config.Config
	logger   *logrus.Logger
	store    *FileStore
	composer *handler.StoreComposer
	handler  *handler.Handler
}

type FileStore struct {
	basePath string
	logger   *logrus.Logger
	config   *config.Config
}

type FileMeta struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	MimeType string    `json:"mime_type"`
	SHA256   string    `json:"sha256"`
	UploadID string    `json:"upload_id"`
	Created  time.Time `json:"created"`
	Device   string    `json:"device"`
}

func NewTusHandler(cfg *config.Config, logger *logrus.Logger) (*TusHandler, error) {
	store := &FileStore{
		basePath: cfg.Storage.UploadDir,
		logger:   logger,
		config:   cfg,
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.Storage.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	composer := handler.NewStoreComposer()
	store.useIn(composer)

	maxSize, err := cfg.GetMaxFileSizeBytes()
	if err != nil {
		logger.WithError(err).Warn("Invalid max file size, using default 10GB")
		maxSize = 10 * 1024 * 1024 * 1024
	}

	tusHandler, err := handler.NewHandler(handler.Config{
		StoreComposer:           composer,
		BasePath:                cfg.TUS.BasePath,
		MaxSize:                 maxSize,
		NotifyCompleteUploads:   true,
		NotifyTerminatedUploads: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tus handler: %w", err)
	}

	return &TusHandler{
		config:   cfg,
		logger:   logger,
		store:    store,
		composer: composer,
		handler:  tusHandler,
	}, nil
}

func (h *TusHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, HEAD, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Upload-Length, Upload-Offset, Tus-Resumable, Upload-Metadata")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Log upload requests
	h.logger.WithFields(logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"remote": r.RemoteAddr,
	}).Info("TUS upload request")

	// Use a custom HTTP handler to ensure proper routing
	h.handler.ServeHTTP(w, r)
}

func (s *FileStore) useIn(composer *handler.StoreComposer) {
	composer.UseCore(s)
	composer.UseTerminater(s)
	composer.UseConcater(s)
	composer.UseLengthDeferrer(s)
}

func (s *FileStore) NewUpload(ctx context.Context, info handler.FileInfo) (handler.Upload, error) {
	// Generate unique file ID
	fileID := uuid.New().String()

	// Extract filename from metadata
	var fileName string
	if metadata, ok := info.MetaData["filename"]; ok {
		fileName = metadata
	} else {
		fileName = fileID
	}

	// Create file path
	filePath := filepath.Join(s.basePath, fileID+s.config.TUS.TempSuffix)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload file: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"upload_id": fileID,
		"filename":  fileName,
		"size":      info.Size,
	}).Info("New upload created")

	return &FileUpload{
		id:        fileID,
		file:      file,
		filePath:  filePath,
		fileName:  fileName,
		size:      info.Size,
		offset:    0,
		info:      info,
		store:     s,
		createdAt: time.Now(),
	}, nil
}

func (s *FileStore) GetUpload(ctx context.Context, id string) (handler.Upload, error) {
	filePath := filepath.Join(s.basePath, id+s.config.TUS.TempSuffix)

	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, handler.ErrNotFound
	}

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileUpload{
		id:       id,
		file:     file,
		filePath: filePath,
		size:     -1, // Unknown size
		offset:   fileInfo.Size(),
		store:    s,
	}, nil
}

func (s *FileStore) AsTerminatableUpload(upload handler.Upload) handler.TerminatableUpload {
	return upload.(*FileUpload)
}

func (s *FileStore) AsLengthDeclarableUpload(upload handler.Upload) handler.LengthDeclarableUpload {
	return upload.(*FileUpload)
}

func (s *FileStore) AsConcatableUpload(upload handler.Upload) handler.ConcatableUpload {
	return upload.(*FileUpload)
}

type FileUpload struct {
	id        string
	file      *os.File
	filePath  string
	fileName  string
	size      int64
	offset    int64
	info      handler.FileInfo
	store     *FileStore
	createdAt time.Time
	mu        sync.RWMutex
}

func (u *FileUpload) WriteChunk(ctx context.Context, offset int64, src io.Reader) (int64, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if offset != u.offset {
		_, err := u.file.Seek(offset, io.SeekStart)
		if err != nil {
			return 0, fmt.Errorf("failed to seek to offset %d: %w", offset, err)
		}
	}

	n, err := io.Copy(u.file, src)
	if err != nil {
		return n, fmt.Errorf("failed to write chunk: %w", err)
	}

	u.offset = offset + n

	u.store.logger.WithFields(logrus.Fields{
		"upload_id":  u.id,
		"offset":     u.offset,
		"chunk_size": n,
	}).Debug("Chunk written")

	return n, nil
}

func (u *FileUpload) GetInfo(ctx context.Context) (handler.FileInfo, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	info := u.info
	info.Offset = u.offset
	info.SizeIsDeferred = u.size < 0
	return info, nil
}

func (u *FileUpload) GetReader(ctx context.Context) (io.ReadCloser, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	file, err := os.Open(u.filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (u *FileUpload) FinishUpload(ctx context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Close the file
	if err := u.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Calculate SHA256
	hash, err := u.calculateSHA256()
	if err != nil {
		return fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	// Move to final location
	finalPath := filepath.Join(u.store.basePath, u.id)
	if err := os.Rename(u.filePath, finalPath); err != nil {
		return fmt.Errorf("failed to move file to final location: %w", err)
	}

	// Create metadata file
	meta := FileMeta{
		ID:       u.id,
		Name:     u.fileName,
		Size:     u.offset,
		MimeType: u.getMimeType(),
		SHA256:   hash,
		UploadID: u.id,
		Created:  u.createdAt,
		Device:   u.getDeviceFromMeta(),
	}

	if err := u.saveMetadata(meta); err != nil {
		u.store.logger.WithError(err).Error("Failed to save file metadata")
		// Continue anyway - the upload is complete
	}

	u.store.logger.WithFields(logrus.Fields{
		"upload_id": u.id,
		"filename":  u.fileName,
		"size":      u.offset,
		"sha256":    hash,
	}).Info("Upload completed")

	return nil
}

func (u *FileUpload) Terminate(ctx context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.file.Close()

	// Remove the partial file
	if err := os.Remove(u.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove upload file: %w", err)
	}

	// Remove metadata file if it exists
	metaPath := filepath.Join(u.store.basePath, u.id+u.store.config.TUS.MetaSuffix)
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		u.store.logger.WithError(err).Warn("Failed to remove metadata file")
	}

	u.store.logger.WithFields(logrus.Fields{
		"upload_id": u.id,
	}).Info("Upload terminated")

	return nil
}

func (u *FileUpload) ConcatUploads(ctx context.Context, uploads []handler.Upload) error {
	// Implementation for file concatenation if needed
	return handler.ErrNotImplemented
}

func (u *FileUpload) DeclareLength(ctx context.Context, length int64) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.size = length
	return nil
}

func (u *FileUpload) calculateSHA256() (string, error) {
	file, err := os.Open(u.filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (u *FileUpload) getMimeType() string {
	// Try to determine MIME type from file extension
	ext := strings.ToLower(filepath.Ext(u.fileName))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

func (u *FileUpload) getDeviceFromMeta() string {
	if deviceMeta, ok := u.info.MetaData["device"]; ok {
		return deviceMeta
	}
	return "unknown"
}

func (u *FileUpload) saveMetadata(meta FileMeta) error {
	metaPath := filepath.Join(u.store.basePath, u.id+u.store.config.TUS.MetaSuffix)

	file, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(meta)
}
