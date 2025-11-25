package domain

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Job struct {
	TaskType string          `json:"task_type"`
	Payload  json.RawMessage `json:"payload"`
}

type ImageJobPayload struct {
	ImageJobID string `json:"imageJobID"`
}

type ReportDoc struct {
	ID        primitive.ObjectID     `bson:"_id"`
	Type      string                 `bson:"type"`
	Status    string                 `bson:"status"`
	FileURL   string                 `bson:"fileURL"`
	ErrorMsg  string                 `bson:"errorMsg"`
	Filters   map[string]interface{} `bson:"filters"`
	CreatedAt primitive.DateTime     `bson:"createdAt"`
	UpdatedAt primitive.DateTime     `bson:"updatedAt"`
}

type ImageJobDoc struct {
	ID             primitive.ObjectID `bson:"_id"`
	Status         string             `bson:"status"`
	SourceImageURL string             `bson:"sourceImageURL"`
	OutputImageURL string             `bson:"outputImageURL"`
	ErrorMsg       string             `bson:"errorMsg"`
	CreatedAt      time.Time          `bson:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt"`
}
