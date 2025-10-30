package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TeamMember struct {
	ID        primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	TeamId string `bson:"teamId" json:"teamId"`
	User string `bson:"user" json:"user"`
	Role      string    `bson:"role" json:"role"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
