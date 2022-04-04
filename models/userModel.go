package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id"`
	FirstName    *string            `json:"first_name" validate:"required,min=2,max=100"`
	LastName     *string            `json:"last_name" validate:"required,min=2,max=100"`
	Password     *string            `json:"password" validate:"required,min=2,max=100"`
	Email        *string            `json:"email" validate:"required,email"`
	Phone        *string            `json:"phone" validate:"required,min=2,max=15"`
	Token        *string            `json:"token"`
	UserType     *string            `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	RefreshToken *string            `bson:"refresh_token" json:"refresh_token"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	UserId       string             `bson:"user_id" json:"user_id"`
}
