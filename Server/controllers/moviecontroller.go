package controllers

import (
	database "Magic/Database"
	Models "Magic/Models"
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/gin-gonic/gin"
)

var moviecollection *mongo.Collection = database.OpenCollection("movies")
var validate = validator.New()

func GetMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var movies []Models.Movie

		cursor, err := moviecollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, movies)
	}
}
func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		movieId := c.Param("imdb_id")
		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "movieeid cannot be an empty string "})
			return

		}
		var movie Models.Movie
		err := moviecollection.FindOne(ctx, bson.M{"imdb_id": movieId}).Decode(&movie)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "failled to decode the movie"})
			return
		}
		c.JSON(http.StatusOK, movie)

	}
}
func Addmovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var movie Models.Movie
		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "an error just occured"})
			return
		}
		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "validation error occurred",
				"detail": err.Error(),
			})
			return
		}
		result, err := moviecollection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "an error occured ", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, result)

	}
}
