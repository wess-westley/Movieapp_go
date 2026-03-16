package controllers

import (
	database "Magic/Database"
	Models "Magic/Models"
	"Magic/utilis"
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

var usercollection *mongo.Collection = database.OpenCollection("users")
var Validate = validator.New()

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user Models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error occurred during registration", "details": err.Error()})
			return
		}

		if err := Validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation error", "details": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		count, err := usercollection.CountDocuments(ctx, bson.M{"Email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing users"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}

		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
			return
		}

		user.Password = hashedPassword
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		result, err := usercollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}
func UserLogin() gin.HandlerFunc {
	return func(c *gin.Context) {

		var userLogin Models.Userlogin

		if err := c.ShouldBindJSON(&userLogin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid input",
			})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var foundUser Models.User

		err := usercollection.FindOne(ctx, bson.M{"Email": userLogin.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user not authenticated",
			})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userLogin.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid password",
			})
			return
		}

		token, refreshToken, err := utilis.GenerateTokens(
			foundUser.FirstName,
			foundUser.LastName,
			foundUser.Email,
			foundUser.UserID,
			foundUser.Role,
			foundUser.MiddleName,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "could not generate auth token",
				"details": err.Error(),
			})
			return
		}

		err = utilis.UpdateTokens(foundUser.UserID, token, refreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "could not update tokens",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Models.Userresponse{
			UserId:        foundUser.UserID,
			FirstName:     foundUser.FirstName,
			MiddleName:    foundUser.MiddleName,
			LastName:      foundUser.LastName,
			Email:         foundUser.Email,
			Role:          foundUser.Role,
			Token:         token,
			Refresh_token: refreshToken,
			Favorite:      foundUser.Favorite,
		})
	}
}
