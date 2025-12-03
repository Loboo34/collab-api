package utils

import (
	"context"
	"time"

	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/services"
)

func Log( userID, teamID, projectID, taskID, action, message string) {
	

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)


	go func () {
		defer cancel()
		err := services.CreateLog(ctx, models.ActivityLog{
		UserID:    userID,
		TeamID:    teamID,
		ProjectID: projectID,
		TaskID:    taskID,
		Action:    action,
		Message:   message,
	})
	if err != nil {
		Logger.Warn("Failed to Log Activity")
	}
	
	}()
}
