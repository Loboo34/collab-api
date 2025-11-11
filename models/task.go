package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct{
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title string `bson:"title" json:"title"`
	Description string `bson:"description" json:"descroption"`
	Status string `bson:"status" json:"status"`//pending, inProgress,done
	AssignedTo primitive.ObjectID `bson:"assigned" json:"assigned"`//id of the team member the task is assinged to
	TeamId primitive.ObjectID `bson:"teamid" json:"teamid"`
	ProjectId primitive.ObjectID `bson:"projectId,omitempty" json:"projectid"`
	CreatedAt time.Time `bson:"createdAt" json:"createdat"`
}