package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	MaxFileSize    = 50 * 1024 * 1024 // 10MB
	StorageRoot    = "storage"
	ModelsPath     = "storage/models"
	ThumbnailsPath = "storage/thumbnails"

	// Public URLs for static file serving
	PublicPrefix     = "/static"
	PublicModels     = "/static/models"
	PublicThumbnails = "/static/thumbnails"
)

var (
	AllowedGlbFormats   = []string{".glb", ".gltf"}
	AllowedUsdzFormats  = []string{".usdz"}
	AllowedImageFormats = []string{".png", ".jpg", ".jpeg"}

	ErrFileTooLarge     = errors.New("file size exceeds maximum limit of 10MB")
	ErrInvalidFormat    = errors.New("invalid file format")
	ErrStorageNotExists = errors.New("storage directory does not exist")
)

type StorageService interface {
	SaveGlbModel(file *multipart.FileHeader, modelID uuid.UUID) (string, error)
	SaveUsdzModel(file *multipart.FileHeader, modelID uuid.UUID) (string, error)
	SaveThumbnail(file *multipart.FileHeader, modelID uuid.UUID) (string, error)
	DeleteGlbModel(modelID uuid.UUID) error
	DeleteUsdzModel(modelID uuid.UUID) error
	DeleteThumbnail(modelID uuid.UUID) error
	GetPublicGlbPath(modelID uuid.UUID) string
	GetPublicUsdzPath(modelID uuid.UUID) string
	GetPublicThumbnailPath(modelID uuid.UUID) string
}

type storageService struct{}

func NewStorageService() StorageService {
	ensureStorageDirs()
	return &storageService{}
}

func (s *storageService) SaveGlbModel(file *multipart.FileHeader, modelID uuid.UUID) (string, error) {
	if file.Size > MaxFileSize {
		return "", ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedFormat(ext, AllowedGlbFormats) {
		return "", ErrInvalidFormat
	}

	filename := fmt.Sprintf("%s%s", modelID.String(), ext)
	path := filepath.Join(ModelsPath, "glb", filename)

	if err := saveFile(file, path); err != nil {
		return "", err
	}

	return s.GetPublicGlbPath(modelID), nil
}

func (s *storageService) SaveUsdzModel(file *multipart.FileHeader, modelID uuid.UUID) (string, error) {
	if file.Size > MaxFileSize {
		return "", ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedFormat(ext, AllowedUsdzFormats) {
		return "", ErrInvalidFormat
	}

	filename := fmt.Sprintf("%s%s", modelID.String(), ext)
	path := filepath.Join(ModelsPath, "usdz", filename)

	if err := saveFile(file, path); err != nil {
		return "", err
	}

	return s.GetPublicUsdzPath(modelID), nil
}

func (s *storageService) SaveThumbnail(file *multipart.FileHeader, modelID uuid.UUID) (string, error) {
	if file.Size > MaxFileSize {
		return "", ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedFormat(ext, AllowedImageFormats) {
		return "", ErrInvalidFormat
	}

	filename := fmt.Sprintf("%s%s", modelID.String(), ext)
	path := filepath.Join(ThumbnailsPath, filename)

	if err := saveFile(file, path); err != nil {
		return "", err
	}

	return s.GetPublicThumbnailPath(modelID), nil
}

func (s *storageService) GetPublicGlbPath(modelID uuid.UUID) string {
	return fmt.Sprintf("%s/glb/%s", PublicModels, modelID.String())
}

func (s *storageService) GetPublicUsdzPath(modelID uuid.UUID) string {
	return fmt.Sprintf("%s/usdz/%s", PublicModels, modelID.String())
}

func (s *storageService) GetPublicThumbnailPath(modelID uuid.UUID) string {
	return fmt.Sprintf("%s/%s", PublicThumbnails, modelID.String())
}

func (s *storageService) DeleteGlbModel(modelID uuid.UUID) error {
	return deleteFileWithAnyExt(filepath.Join(ModelsPath, "glb", modelID.String()), AllowedGlbFormats)
}

func (s *storageService) DeleteUsdzModel(modelID uuid.UUID) error {
	return deleteFileWithAnyExt(filepath.Join(ModelsPath, "usdz", modelID.String()), AllowedUsdzFormats)
}

func (s *storageService) DeleteThumbnail(modelID uuid.UUID) error {
	return deleteFileWithAnyExt(filepath.Join(ThumbnailsPath, modelID.String()), AllowedImageFormats)
}

func ensureStorageDirs() {
	os.MkdirAll(filepath.Join(ModelsPath, "glb"), 0755)
	os.MkdirAll(filepath.Join(ModelsPath, "usdz"), 0755)
	os.MkdirAll(ThumbnailsPath, 0755)
}

func saveFile(file *multipart.FileHeader, path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func deleteFileWithAnyExt(basePath string, allowedExts []string) error {
	for _, ext := range allowedExts {
		path := basePath + ext
		if err := os.Remove(path); err == nil {
			return nil
		}
	}
	return os.ErrNotExist
}

func isAllowedFormat(ext string, allowedFormats []string) bool {
	for _, format := range allowedFormats {
		if ext == format {
			return true
		}
	}
	return false
}
