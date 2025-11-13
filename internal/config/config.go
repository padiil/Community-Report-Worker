package config

import (
	"context"
	"log/slog"
	"os"
	"time"

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
	redisURI := os.Getenv("REDIS_URI")
	client := redis.NewClient(&redis.Options{
		Addr: redisURI,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("Gagal terhubung ke Redis", "err", err)
		os.Exit(1)
	}
	slog.Info("Berhasil terhubung ke Redis", "uri", redisURI)
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

func LoadR2Config() *R2Config {
	return &R2Config{
		Endpoint:  os.Getenv("R2_ENDPOINT"),
		AccessKey: os.Getenv("R2_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		Bucket:    os.Getenv("R2_BUCKET_NAME"),
		PublicURL: os.Getenv("R2_PUBLIC_URL"),
	}
}
