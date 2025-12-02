package config

import (
	"context"
	"log/slog"
	"os"
	"time"

	"org-worker/internal/storage"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type R2Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	PublicURL string
}

func GetOrgName() string {
	name := os.Getenv("ORG_NAME")
	if name == "" {
		return "Community Organization"
	}
	return name
}

func InitRedis() *redis.Client {
	ctx := context.Background()
	redisAddr := os.Getenv("REDIS_URI")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("Gagal terhubung ke Redis", "err", err)
		os.Exit(1)
	}
	slog.Info("Berhasil terhubung ke Redis", "uri", redisAddr)
	return client
}

func InitMongo() *mongo.Database {
	ctx := context.Background()
	mongoURI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		slog.Error("Gagal membuat koneksi client Mongo", "err", err)
		os.Exit(1)
	}
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(ctxPing, nil); err != nil {
		slog.Error("Gagal terhubung ke MongoDB", "err", err)
		os.Exit(1)
	}
	slog.Info("Berhasil terhubung ke MongoDB", "uri", mongoURI)
	return client.Database("veritas")
}

func InitStorageProvider(ctx context.Context, logger *slog.Logger) storage.StorageProvider {
	storageType := os.Getenv("STORAGE_PROVIDER")
	if storageType == "r2" {
		logger.Info("Using Cloudflare R2 Storage")
		r2Endpoint := os.Getenv("R2_ENDPOINT")
		r2AccessKey := os.Getenv("R2_ACCESS_KEY_ID")
		r2SecretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
		r2Bucket := os.Getenv("R2_BUCKET_NAME")
		r2PublicURL := os.Getenv("R2_PUBLIC_URL")
		r2Region := os.Getenv("R2_REGION")
		if r2Region == "" {
			r2Region = "auto"
		}
		storageProvider, err := storage.NewR2Storage(ctx, r2Endpoint, r2AccessKey, r2SecretKey, r2Bucket, r2PublicURL, r2Region)
		if err != nil {
			logger.Error("Failed to initialize R2 Storage", "err", err)
			os.Exit(1)
		}
		return storageProvider
	}
	logger.Info("Using Local File Storage")
	return storage.NewLocalStorage("./reports")
}
