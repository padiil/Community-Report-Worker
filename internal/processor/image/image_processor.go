package image

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"org-worker/internal/domain"
	"org-worker/internal/storage"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *ImageHandler) HandleImageProcessing(ctx context.Context, logger *slog.Logger, imageJob domain.ImageJobDoc, storageProvider storage.StorageProvider) error {
	logger.Info("Processing image", "imageJobID", imageJob.ID, "sourceImageURL", imageJob.SourceImageURL)

	src, err := imaging.Open(imageJob.SourceImageURL)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %v", imageJob.SourceImageURL, err)
	}
	resizedImg := imaging.Resize(src, 800, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)
	if err := webp.Encode(buf, resizedImg, &webp.Options{Quality: 80}); err != nil {
		return fmt.Errorf("failed to encode image to webp: %v", err)
	}

	filename := fmt.Sprintf("%s-resized.webp", imageJob.ID)
	imageURL, err := storageProvider.Save(ctx, "images", filename, buf)
	if err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	// Update status & outputImageURL di MongoDB
	update := bson.M{
		"$set": bson.M{
			"status":         "COMPLETED",
			"outputImageURL": imageURL,
			"updatedAt":      primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	err = h.repo.UpdateStatus(ctx, imageJob.ID, update)
	if err != nil {
		logger.Error("Failed to update image job status in MongoDB", "err", err)
		return fmt.Errorf("failed to update image job status: %v", err)
	}

	logger.Info("Image saved", "imageJobID", imageJob.ID, "outputURL", imageURL)

	return nil
}
