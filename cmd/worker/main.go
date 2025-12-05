package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"sync"

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

	maxConcurrency := 10
	if val := os.Getenv("MAX_CONCURRENCY"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxConcurrency = parsed
		}
	}

	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	logger.Info("Worker is running with concurrency limit:", "limit", maxConcurrency)

	for {
		result, err := queue.BLPop(redisClient, ctx, "task_queue")
		if err != nil {
			logger.Error("Error reading from Redis queue", "err", err)
			continue
		}
		if len(result) < 2 {
			continue
		}

		sem <- struct{}{}

		wg.Add(1)

		go func(data string) {
			defer wg.Done()
			defer func() { <-sem }()

			var job domain.Job
			if err := json.Unmarshal([]byte(data), &job); err != nil {
				logger.Error("Failed to unmarshal job", "err", err)
				return
			}

			switch job.TaskType {
			case "generate_report":
				var payload domain.ReportJobPayload
				if err := json.Unmarshal(job.Payload, &payload); err != nil {
					logger.Error("Failed to unmarshal report payload", "err", err)
					return
				}
				logger.Info("Received job", "reportID", payload.ReportID, "taskType", job.TaskType)

				reportDoc, err := reportRepo.GetReportByID(ctx, payload.ReportID)
				if err != nil {
					logger.Error("Failed to get report", "err", err)
					return
				}
				if err := reportHandler.HandleReportGeneration(ctx, logger, reportDoc); err != nil {
					logger.Error("ERROR generate_report", "err", err)
				}

			case "process_image":
				var payload domain.ImageJobPayload
				if err := json.Unmarshal(job.Payload, &payload); err != nil {
					logger.Error("Failed to unmarshal image payload", "err", err)
					return
				}
				logger.Info("Starting job processing", "imageJobID", payload.ImageJobID)

				if err := imageHandler.HandleImageProcessing(ctx, logger, payload.ImageJobID); err != nil {
					logger.Error("ERROR process_image", "err", err)
				}
			default:
				logger.Warn("Unknown task type", "taskType", job.TaskType)
			}
		}(result[1])
	}
}
