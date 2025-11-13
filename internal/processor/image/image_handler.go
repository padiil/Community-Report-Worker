package image

import (
	"org-worker/internal/repository"
	"org-worker/internal/storage"
)

type ImageHandler struct {
	repo    *repository.ImageJobRepository
	storage storage.StorageProvider
}

func NewImageHandler(repo *repository.ImageJobRepository, storage storage.StorageProvider) *ImageHandler {
	return &ImageHandler{repo: repo, storage: storage}
}
