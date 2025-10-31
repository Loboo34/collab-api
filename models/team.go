package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
	
)

type Team struct{
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
	Members []string `bson:"members" json:"members"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	CreatedBy string  `bson:"createdby" json:"createdby"`
}