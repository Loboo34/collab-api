package services

import (
	"context"
	"time"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
)

func CreateLog(ctx context.Context, log models.ActivityLog) error{
	log.Timestamp = time.Now()

	collection := database.DB.Collection("activity-log")

	_, err := collection.InsertOne(ctx, log)

	return err
}