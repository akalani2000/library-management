package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title,omitempty" bson:"title,omitempty"`
	Author      string             `json:"author,omitempty" bson:"author,omitempty"`
	Publisher   string             `json:"publisher,omitempty" bson:"publisher,omitempty"`
	PublishDate string             `json:"publish_date,omitempty" bson:"publish_date,omitempty"`
	ISBN        string             `json:"isbn,omitempty" bson:"isbn,omitempty"`
	CoverImage  string             `json:"cover_image,omitempty" bson:"cover_image,omitempty"`
	BookPDF     string             `json:"book_pdf,omitempty" bson:"book_pdf,omitempty"`
	Tags        []string           `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt   time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type BookRequest struct {
	Title       string   `form:"Title" json:"title,omitempty"`
	Author      string   `form:"Author" json:"author,omitempty"`
	Publisher   string   `form:"Publisher" json:"publisher,omitempty"`
	PublishDate string   `form:"PublishDate" json:"publish_date,omitempty"`
	ISBN        string   `form:"ISBN" json:"isbn,omitempty"`
	Tags        []string `form:"Tags" json:"tags,omitempty"`
}
