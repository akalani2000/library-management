package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole int

const (
	SuperUser UserRole = iota
	ManagerRole
	StudentRole
)

type SystemUser struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name             string             `json:"name,omitempty" bson:"name,omitempty"`
	Email            string             `json:"email,omitempty" bson:"email,omitempty"`
	Password         string             `json:"password,omitempty" bson:"password,omitempty"`
	Token            string             `json:"token,omitempty" bson:"token,omitempty"`
	IsSuperuser      bool               `json:"is_superuser" bson:"is_superuser"`
	Role             UserRole           `json:"role" bson:"role"`
	StripeCustomerID string             `json:"stripe_customer_id,omitempty" bson:"stripe_customer_id,omitempty"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}

type SystemUserRegisterResponse struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Email       string             `json:"email,omitempty" bson:"email,omitempty"`
	IsSuperuser bool               `json:"is_superuser" bson:"is_superuser"`
	Role        UserRole           `json:"role" bson:"role"`
}

type SystemUserLoginResponse struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Email       string             `json:"email,omitempty" bson:"email,omitempty"`
	IsSuperuser bool               `json:"is_superuser" bson:"is_superuser"`
	Role        UserRole           `json:"role" bson:"role"`
	Token       string             `json:"token,omitempty" bson:"token,omitempty"`
}

type SystemUserRegister struct {
	Name     string `json:"name,omitempty" bson:"name,omitempty"`
	Email    string `json:"email,omitempty" bson:"email,omitempty"`
	Password string `json:"password,omitempty" bson:"password,omitempty"`
}

type SystemUserLogin struct {
	Email    string `form:"Email" json:"email,omitempty" bson:"email,omitempty"`
	Password string `form:"Password" json:"password,omitempty" bson:"password,omitempty"`
}
