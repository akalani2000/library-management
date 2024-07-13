package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionType integer choice field
type SubscriptionType int

const (
	NoRecurring SubscriptionType = iota
	Monthly
	Quarterly
	Yearly
)

type Subscription struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Type        SubscriptionType   `json:"type" bson:"type"`
	Price       float64            `json:"price" bson:"price"`
	ProductID   string             `json:"product_id" bson:"product_id"`
	PriceID     string             `json:"price_id" bson:"price_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type SubscriptionCreateRequest struct {
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description" binding:"required"`
	Type        SubscriptionType `json:"type" binding:"required"`
	Price       float64          `json:"price" binding:"required"`
}

type SubscriptionUpdateRequest struct {
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
	Type        SubscriptionType `json:"type,omitempty"`
	Price       float64          `json:"price,omitempty"`
}

// SubscriptionStatus integer choice field
type SubscriptionStatus int

const (
	Pending SubscriptionStatus = iota
	Subscribed
	Cancelled
	Expired
	InRecurring
)

type PaymentStatus int

const (
	Open PaymentStatus = iota
	Paid
	Failed
)

// StudentSubscription model
type StudentSubscription struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SystemUserID   primitive.ObjectID `bson:"system_user_id" json:"system_user_id"`
	SubscriptionID primitive.ObjectID `bson:"subscription_id" json:"subscription_id"`
	CustomerID     string             `bson:"customer_id" json:"customer_id"`
	PriceID        string             `bson:"price_id" json:"price_id"`
	StripeSubID    string             `bson:"stripe_sub_id" json:"stripe_sub_id"`
	InvoiceID      string             `bson:"invoice_id" json:"invoice_id"`
	PaymentLink    string             `bson:"payment_link" json:"payment_link"`
	PaymentStatus  PaymentStatus      `bson:"payment_status" json:"payment_status"`
	Status         SubscriptionStatus `bson:"status" json:"status"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type StudentSubscriptionRequest struct {
	SubscriptionID primitive.ObjectID `bson:"subscription_id" json:"subscription_id"`
}
