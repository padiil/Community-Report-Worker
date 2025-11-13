package repository

import (
	"context"
	"org-worker/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ImageJobRepository struct {
	coll *mongo.Collection
}

func NewImageJobRepository(db *mongo.Database) *ImageJobRepository {
	return &ImageJobRepository{coll: db.Collection("image_jobs")}
}

func (r *ImageJobRepository) GetImageJobByID(ctx context.Context, id string) (domain.ImageJobDoc, error) {
	var doc domain.ImageJobDoc
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return doc, err
	}
	err = r.coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&doc)
	return doc, err
}

func (r *ImageJobRepository) UpdateStatus(ctx context.Context, id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.UpdateByID(ctx, objID, update)
	return err
}
