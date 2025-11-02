package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Invite struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamID    primitive.ObjectID `bson:"teamId" json:"teamId"`
	Email     string             `bson:"email" json:"email"`
	Token     string             `bson:"token" json:"token"`
	Status    string             `bson:"status" json:"status"` // pending, accepted, declined
	SentBy    string             `bson:"sentBy" json:"sentBy"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
