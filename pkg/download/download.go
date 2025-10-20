package download

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/easy-sync/easy-sync/pkg/config"
)

type Handler struct {
	config *config.Config
	logger *logrus.Logger
}

type FileMetadata struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	MimeType string    `json:"mime_type"`
	SHA256   string    `json:"sha256"`
	UploadID string    `json:"upload_id"`
	Created  time.Time `json:"created"`
	Device   string    `json:"device"`
}

func NewHandler(cfg *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{
		config: cfg,
		logger: logger,
	}
}

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Range, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	fileID := strings.TrimPrefix(r.URL.Path, "/files/")
	if fileID == "" {
		http.Error(w, "File ID required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.config.Storage.UploadDir, fileID)
	metaPath := filepath.Join(h.config.Storage.UploadDir, fileID+".meta")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Load metadata
	meta, err := h.loadMetadata(metaPath)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to load file metadata")
		// Continue without metadata
		meta = &FileMetadata{
			ID:   fileID,
			Name: fileID,
			Size: h.getFileSize(filePath),
		}
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	// Set content type
	if meta.MimeType != "" {
		w.Header().Set("Content-Type", meta.MimeType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Set content disposition
	fileName := meta.Name
	if fileName == "" {
		fileName = fileID
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	// Set content length
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// Handle range requests
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		h.handleRangeRequest(w, r, filePath, fileInfo.Size())
		return
	}

	// Serve file
	h.serveFile(w, r, filePath, meta)
}

func (h *Handler) HandleSHA256(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	fileID := strings.TrimPrefix(r.URL.Path, "/files/")
	fileID = strings.TrimSuffix(fileID, "/sha256")

	if fileID == "" {
		http.Error(w, "File ID required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.config.Storage.UploadDir, fileID)
	metaPath := filepath.Join(h.config.Storage.UploadDir, fileID+".meta")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Try to get SHA256 from metadata first
	meta, err := h.loadMetadata(metaPath)
	if err == nil && meta.SHA256 != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(meta.SHA256))
		return
	}

	// Calculate SHA256 if not in metadata
	sha256, err := h.calculateSHA256(filePath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to calculate SHA256")
		http.Error(w, "Failed to calculate checksum", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(sha256))
}

func (h *Handler) handleRangeRequest(w http.ResponseWriter, r *http.Request, filePath string, fileSize int64) {
	rangeHeader := r.Header.Get("Range")

	// Parse range header
	ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if len(ranges) != 2 {
		http.Error(w, "Invalid range header", http.StatusBadRequest)
		return
	}

	var start, end int64
	var err error

	if ranges[0] != "" {
		start, err = strconv.ParseInt(ranges[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid range start", http.StatusBadRequest)
			return
		}
	}

	if ranges[1] != "" {
		end, err = strconv.ParseInt(ranges[1], 10, 64)
		if err != nil {
			http.Error(w, "Invalid range end", http.StatusBadRequest)
			return
		}
	} else {
		end = fileSize - 1
	}

	// Validate range
	if start < 0 || end >= fileSize || start > end {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Set range headers
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Serve range
	h.serveFileRange(w, filePath, start, end)
}

func (h *Handler) serveFile(w http.ResponseWriter, r *http.Request, filePath string, meta *FileMetadata) {
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Use sendfile for efficient copying if available
	// For simplicity, we'll use io.Copy which is still efficient
	_, err = io.Copy(w, file)
	if err != nil {
		h.logger.WithError(err).Error("Failed to serve file")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"file_id":   meta.ID,
		"file_name": meta.Name,
		"size":      meta.Size,
		"remote":    r.RemoteAddr,
	}).Info("File downloaded")
}

func (h *Handler) serveFileRange(w http.ResponseWriter, filePath string, start, end int64) {
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Seek to start position
	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		http.Error(w, "Failed to seek file", http.StatusInternalServerError)
		return
	}

	// Copy specified range
	limitedReader := io.LimitReader(file, end-start+1)
	_, err = io.Copy(w, limitedReader)
	if err != nil {
		h.logger.WithError(err).Error("Failed to serve file range")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"start": start,
		"end":   end,
		"size":  end - start + 1,
	}).Debug("File range served")
}

func (h *Handler) loadMetadata(metaPath string) (*FileMetadata, error) {
	file, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var meta FileMetadata
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (h *Handler) getFileSize(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

func (h *Handler) calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
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

func (h *Handler) ListFiles() ([]*FileMetadata, error) {
	files := make([]*FileMetadata, 0)

	entries, err := os.ReadDir(h.config.Storage.UploadDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip metadata files
		if strings.HasSuffix(entry.Name(), ".meta") {
			continue
		}

		// Try to load metadata
		metaPath := filepath.Join(h.config.Storage.UploadDir, entry.Name()+".meta")
		meta, err := h.loadMetadata(metaPath)
		if err != nil {
			// Create minimal metadata from file info
			filePath := filepath.Join(h.config.Storage.UploadDir, entry.Name())
			meta = &FileMetadata{
				ID:   entry.Name(),
				Name: entry.Name(),
				Size: h.getFileSize(filePath),
			}
		}

		files = append(files, meta)
	}

	return files, nil
}

func (h *Handler) DeleteFile(fileID string) error {
	filePath := filepath.Join(h.config.Storage.UploadDir, fileID)
	metaPath := filepath.Join(h.config.Storage.UploadDir, fileID+".meta")

	// Delete file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete metadata
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		h.logger.WithError(err).Warn("Failed to delete metadata file")
	}

	h.logger.WithField("file_id", fileID).Info("File deleted")
	return nil
}

func (h *Handler) GenerateDownloadURL(fileID string, expires time.Duration) (string, error) {
	// In a real implementation, you would generate a signed URL with expiration
	// For now, return a simple URL
	return fmt.Sprintf("http://%s/files/%s", h.config.GetAddr(), fileID), nil
}
