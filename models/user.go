package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	FulllName string    `bson:"fullname" json:"fullname"`
	Email     string    `bson:"email" json:"email"`
	Password  string    `bson:"password,omitempty" json:"password,omitempty"`
	Teams     []primitive.ObjectID  `bson:"teams" json:"teams"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
