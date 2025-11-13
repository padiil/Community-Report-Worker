package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"org-worker/internal/config"
	"org-worker/internal/domain"
	"org-worker/internal/processor/image"
	"org-worker/internal/processor/report"
	"org-worker/internal/queue"
	"org-worker/internal/repository"
	"org-worker/internal/storage"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	godotenv.Load(".env")
	redisClient := config.InitRedis()
	mongoDB := config.InitMongo()
	reportRepo := repository.NewReportRepository(mongoDB)
	imageJobRepo := repository.NewImageJobRepository(mongoDB) // Tambahkan repository image jobs

	// --- Setup Storage Provider ---
	var storageProvider storage.StorageProvider
	storageType := os.Getenv("STORAGE_PROVIDER")
	ctx := context.Background()
	if storageType == "r2" {
		logger.Info("Using Cloudflare R2 Storage")
		r2Cfg := config.LoadR2Config()
		var err error
		storageProvider, err = storage.NewR2Storage(ctx, r2Cfg.Endpoint, r2Cfg.AccessKey, r2Cfg.SecretKey, r2Cfg.Bucket, r2Cfg.PublicURL)
		if err != nil {
			logger.Error("Failed to initialize R2 Storage", "err", err)
			os.Exit(1)
		}
	} else {
		logger.Info("Using Local File Storage")
		storageProvider = storage.NewLocalStorage("./reports")
	}

	reportHandler := report.NewReportHandler(reportRepo, storageProvider)
	imageHandler := image.NewImageHandler(imageJobRepo, storageProvider) // Handler baru untuk image jobs

	logger.Info("Worker is running...")
	for {
		result, err := queue.BLPop(redisClient, ctx, "task_queue")
		if err != nil {
			logger.Error("Error reading from Redis queue", "err", err)
			continue
		}
		if len(result) < 2 {
			logger.Warn("Received empty job from queue")
			continue
		}
		var job domain.Job
		if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
			logger.Error("Failed to unmarshal job", "err", err)
			continue
		}

		switch job.TaskType {
		case "generate_report":
			var payload domain.ReportJobPayload
			if err := json.Unmarshal(job.Payload, &payload); err != nil {
				logger.Error("Failed to unmarshal report payload", "err", err)
				continue
			}
			logger.Info("Received job", "reportID", payload.ReportID, "taskType", job.TaskType)
			// Ambil detail report dari MongoDB
			reportDoc, err := reportRepo.GetReportByID(ctx, payload.ReportID)
			if err != nil {
				logger.Error("Failed to get report from MongoDB", "err", err)
				continue
			}
			if err := reportHandler.HandleReportGeneration(ctx, logger, reportDoc); err != nil {
				logger.Error("ERROR generate_report", "err", err)
			}
		case "process_image":
			var payload domain.ImageJobPayload
			if err := json.Unmarshal(job.Payload, &payload); err != nil {
				logger.Error("Failed to unmarshal image payload", "err", err)
				continue
			}
			logger.Info("Received job", "imageJobID", payload.ImageJobID, "taskType", job.TaskType)
			// Ambil detail image job dari MongoDB
			imageJobDoc, err := imageJobRepo.GetImageJobByID(ctx, payload.ImageJobID)
			if err != nil {
				logger.Error("Failed to get image job from MongoDB", "err", err)
				continue
			}
			if err := imageHandler.HandleImageProcessing(ctx, logger, imageJobDoc, storageProvider); err != nil {
				logger.Error("ERROR process_image", "err", err)
			}
		default:
			logger.Warn("Unknown task type", "taskType", job.TaskType)
		}
	}
}
