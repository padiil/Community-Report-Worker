package image

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"org-worker/internal/repository"
	"org-worker/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ImageHandler struct {
	repo    *repository.ImageJobRepository
	storage storage.StorageProvider
}

func NewImageHandler(repo *repository.ImageJobRepository, storage storage.StorageProvider) *ImageHandler {
	return &ImageHandler{repo: repo, storage: storage}
}

func (h *ImageHandler) HandleImageProcessing(ctx context.Context, logger *slog.Logger, jobID string) error {

	// 1. Ambil data Job terbaru dari DB
	imageJob, err := h.repo.FindByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to find job: %v", err)
	}

	logger.Info("Processing image", "jobID", jobID, "source", imageJob.SourceImageURL)

	// 2. Download gambar mentah
	resp, err := http.Get(imageJob.SourceImageURL)
	if err != nil {
		h.handleError(ctx, imageJob.ID, "failed to download request: "+err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("failed to download image, status: %d", resp.StatusCode)
		h.handleError(ctx, imageJob.ID, errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	// 3. Proses Gambar (Stream -> Memory -> Stream)
	webpBuf, err := ProcessImage(resp.Body)
	if err != nil {
		h.handleError(ctx, imageJob.ID, "failed to process image: "+err.Error())
		return err
	}

	// 4. Tentukan nama file
	origName := filepath.Base(imageJob.SourceImageURL)
	baseName := strings.TrimSuffix(origName, filepath.Ext(origName))
	newFilename := baseName + "-optimized.webp"

	// 5. Upload ke R2 (Gunakan h.storage, bukan parameter luar)
	imageURL, err := h.storage.Save(ctx, "optimized", newFilename, webpBuf)
	if err != nil {
		h.handleError(ctx, imageJob.ID, "failed to upload image: "+err.Error())
		return err
	}

	// 6. Update Sukses
	updateData := bson.M{
		"status":         "completed",
		"outputImageURL": imageURL,
		"errorMsg":       "",
	}

	if err := h.repo.UpdateStatus(ctx, imageJob.ID, updateData); err != nil {
		logger.Error("Failed to update success status", "err", err)
		return err
	}

	logger.Info("Image processing completed", "jobID", jobID, "url", imageURL)
	return nil
}

func (h *ImageHandler) handleError(ctx context.Context, id primitive.ObjectID, msg string) {
	h.repo.UpdateStatus(ctx, id, bson.M{
		"status":   "failed",
		"errorMsg": msg,
	})
}
