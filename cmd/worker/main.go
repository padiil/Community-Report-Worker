package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"org-worker/internal/config"
	"org-worker/internal/domain"
	"org-worker/internal/processor/image"
	"org-worker/internal/processor/report"
	"org-worker/internal/queue"
	"org-worker/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	godotenv.Load(".env")
	mongoDB := config.InitMongo()
	redisClient := config.InitRedis()
	reportRepo := repository.NewReportRepository(mongoDB)
	imageJobRepo := repository.NewImageJobRepository(mongoDB)

	ctx := context.Background()
	storageProvider := config.InitStorageProvider(ctx, logger)

	reportHandler := report.NewReportHandler(reportRepo, storageProvider)
	imageHandler := image.NewImageHandler(imageJobRepo, storageProvider)

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
			if err := imageHandler.HandleImageProcessing(ctx, logger, payload.ImageJobID); err != nil {
				logger.Error("ERROR process_image", "err", err)
			}
		default:
			logger.Warn("Unknown task type", "taskType", job.TaskType)
		}
	}
}
