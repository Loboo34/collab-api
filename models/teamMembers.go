package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TeamMember struct {
	ID        primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	TeamId primitive.ObjectID `bson:"teamId" json:"teamId"`
	User string `bson:"user" json:"user"`
	Role      string    `bson:"role" json:"role"`
	JoinedAt time.Time `bson:"joinedat" json:"joinedat"`
}
