package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Project struct{
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
	Description string `bson:"description" json:"descroption"`
	TeamId primitive.ObjectID `bson:"teamId" json:"teamId"`
	CreatedBy string `bson:"createdBy" json:"createdBy"`
	CreatedAt time.Time `bson:"createdAt" json:"createdat"`
}