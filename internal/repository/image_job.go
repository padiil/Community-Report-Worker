package repository

import (
	"context"
	"fmt"
	"time"

	"org-worker/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ImageJobRepository struct {
	collection *mongo.Collection
}

func NewImageJobRepository(db *mongo.Database) *ImageJobRepository {
	return &ImageJobRepository{
		collection: db.Collection("image_jobs"),
	}
}

func (r *ImageJobRepository) FindByID(ctx context.Context, id string) (*domain.ImageJobDoc, error) {
	// 1. Konversi string ID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object id: %v", err)
	}

	// 2. Cari di database
	var job domain.ImageJobDoc
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("image job not found")
		}
		return nil, err
	}

	return &job, nil
}

func (r *ImageJobRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
	fields["updatedAt"] = time.Now()

	update := bson.M{
		"$set": fields,
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update image job: %v", err)
	}
	return nil
}
