package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityLog struct {
	ID        primitive.ObjectID `bson:"_id, omitempty"`
	UserID    string             `bson:"userID"`
	TeamID    string             `bson:"teamID, omitempty"`
	ProjectID string             `bson:"projectID, omitempty"`
	TaskID    string             `bson:"taskID, omitempty"`
	Action    string             `bson:"action"`
	Message   string             `bson:"message"`
	Timestamp time.Time          `bson:"timestamp"`
}
