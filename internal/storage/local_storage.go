package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalFileStorage struct {
	BasePath string
}

func NewLocalStorage(basePath string) *LocalFileStorage {
	return &LocalFileStorage{BasePath: basePath}
}

func (s *LocalFileStorage) Save(ctx context.Context, reportType, filename string, file *bytes.Buffer) (string, error) {
	folderPath := filepath.Join(s.BasePath, reportType)
	fullPath := filepath.Join(folderPath, filename)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", fmt.Errorf("gagal buat folder: %w", err)
	}
	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("gagal buat file: %w", err)
	}
	defer outFile.Close()
	if _, err := io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("gagal salin file: %w", err)
	}
	return fullPath, nil
}
