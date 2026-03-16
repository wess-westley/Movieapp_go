package utilis

import (
	database "Magic/Database"
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

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

	claims := &SignedDetails{
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshClaims := &SignedDetails{
		Email:      email,
		Role:       role,
		FirstName:  firstname,
		LastName:   lastname,
		MiddleName: middlename,
		UserId:     userid,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Magic",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	signedRefreshToken, err := refreshToken.SignedString([]byte(SECRET_REFRESHKEY))
	if err != nil {
		return "", "", err
	}

	return signedToken, signedRefreshToken, nil
}

func UpdateTokens(userId, token, refreshToken string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	updatedData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			"updated_at":    time.Now(),
		},
	}

	_, err := usercollection.UpdateOne(
		ctx,
		bson.M{"user_id": userId},
		updatedData,
	)

	if err != nil {
		return err
	}

	return nil
}
func GetAccessToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("invalid authorization format")
	}

	tokenString := strings.TrimPrefix(authHeader, prefix)
	if tokenString == "" {
		return "", errors.New("bearer token is required")
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*SignedDetails, error) {
	claims := &SignedDetails{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil

}
func GetUserIdFromContext(c *gin.Context) (string, error) {
	userId, exists := c.Get("userId")
	if !exists {
		return "", errors.New("userId doesn't exists")

	}
	id, ok := userId.(string)
	if !ok {
		return "", errors.New("userId doesn't exists hence failed to retrieve")

	}
	return id, nil

}
