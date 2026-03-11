package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`

	FirstName string `bson:"FirstName" json:"FirstName" validate:"required,max=100,min=3"`

	MiddleName string `bson:"MiddleName" json:"MiddleName" validate:"required,max=100,min=3"`

	LastName string `bson:"Lastname" json:"Lastname" validate:"required,max=100,min=4"`

	Email string `bson:"Email" json:"Email" validate:"required,email"`

	Password     string    `bson:"password" json:"password" validate:"required,min=5"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
	Role         string    `bson:"role" json:"role" validate:"oneof=ADMIN USER"`
	Token        string    `bson:"token" json:"token"`
	RefreshToken string    `bson:"refresh_token" json:"refresh_token"`
	Favorite     []Genre   `bson:"favorite" json:"favorite" validate:"required,dive"`
}
type Userlogin struct {
	Email    string `json:"Email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
type Userresponse struct {
	FirstName  string `json:"FirstName"`
	MiddleName string `json:"MiddleName"`
	Email      string `json:"Email"`
	Role       string `json:"Role"`

	LastName string  `json:"Lastname" `
	Favorite []Genre `json:"favorite_genres" `
}
