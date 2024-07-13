package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Manager struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	SystemUserID primitive.ObjectID `json:"system_user_id" bson:"system_user_id"`
	FirstName    string             `json:"first_name" bson:"first_name"`
	LastName     string             `json:"last_name" bson:"last_name"`
	Email        string             `json:"email" bson:"email"`
	ManagerID    string             `json:"manager_id" bson:"manager_id"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

type ManagerRegister struct {
	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`
	Email     string `json:"email" bson:"email"`
	Password  string `json:"password,omitempty" bson:"password,omitempty"`
	ManagerID string `json:"manager_id" bson:"manager_id"`
}

type ManagerUpdate struct {
	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`
	Email     string `json:"email" bson:"email"`
	ManagerID string `json:"manager_id" bson:"manager_id"`
}
