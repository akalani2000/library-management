package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Student struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	SystemUserID   primitive.ObjectID   `bson:"system_user_id" json:"system_user_id"`
	FirstName      string               `bson:"first_name" json:"first_name"`
	LastName       string               `bson:"last_name" json:"last_name"`
	Email          string               `bson:"email" json:"email"`
	StudentID      string               `bson:"student_id" json:"student_id"`
	SubscriptionID primitive.ObjectID   `bson:"subscription_id" json:"subscription_id"`
	Books          []primitive.ObjectID `bson:"books" json:"books"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at" json:"updated_at"`
}

type StudentRegister struct {
	FirstName string `bson:"first_name" json:"first_name"`
	LastName  string `bson:"last_name" json:"last_name"`
	Email     string `bson:"email" json:"email"`
	StudentID string `bson:"student_id" json:"student_id"`
	Password  string `bson:"password,omitempty" json:"password,omitempty" `
}

type StudentUpdate struct {
	FirstName string `bson:"first_name" json:"first_name"`
	LastName  string `bson:"last_name" json:"last_name"`
	Email     string `bson:"email" json:"email"`
	StudentID string `bson:"student_id" json:"student_id"`
}
