package utils

import (
	"context"

	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/services"
)

func Log(ctx context.Context, userID, teamID, projectID, taskID, action, message string) {
	_ = services.CreateLog(ctx, models.ActivityLog{
		UserID:    userID,
		TeamID:    teamID,
		ProjectID: projectID,
		TaskID:    taskID,
		Action:    action,
		Message:   message,
	})
}
