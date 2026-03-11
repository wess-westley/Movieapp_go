package utilis

import (
	database "Magic/Database"
	"context"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SignedDetails struct {
	FirstName  string
	LastName   string
	Email      string
	MiddleName string
	Role       string
	UserId     string
	jwt.RegisteredClaims
}

var SECRET_KEY string = os.Getenv("SECRET_KEY")
var SECRET_REFRESHKEY string = os.Getenv("SECRET_REFRESHKEY")
var usercollection *mongo.Collection = database.OpenCollection("users")

func GenerateTokens(firstname, lastname, email, userid, role, middlename string) (string, string, error) {
	Claims := &SignedDetails{
		Email:      email,
		Role:       role,
		FirstName:  firstname,
		LastName:   lastname,
		MiddleName: middlename,
		UserId:     userid,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Magic",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, Claims)
	signedtoken, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	Refreshclaims := &SignedDetails{
		Email:      email,
		Role:       role,
		FirstName:  firstname,
		LastName:   lastname,
		MiddleName: middlename,
		UserId:     userid,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Magic",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	RefreshToken := jwt.NewWithClaims(jwt.SigningMethodES256, Refreshclaims)
	Refreshsignedtoken, err := RefreshToken.SignedString([]byte(SECRET_REFRESHKEY))
	if err != nil {
		return "", "", err
	}
	return signedtoken, Refreshsignedtoken, nil
}
func UpdateTokens(UserId, RefreshToken, token string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	UpdatedAt := time.Now()

	updatedData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": RefreshToken,
			"updated_at":    UpdatedAt,
		},
	}

	_, err := usercollection.UpdateOne(
		ctx,
		bson.M{"user_id": UserId},
		updatedData,
	)

	if err != nil {
		return err
	}

	return nil
}
