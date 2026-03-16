package controllers

import (
	database "Magic/Database"
	Models "Magic/Models"
	"Magic/utilis"
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var moviecollection *mongo.Collection = database.OpenCollection("movies")
var rankingscollection *mongo.Collection = database.OpenCollection("rankings")

/*var usercollection *mongo.Collection = database.OpenCollection("users")*/

var validate = validator.New()

// GET ALL MOVIES

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

// GET SINGLE MOVIE

func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieId := c.Param("imdb_id")

		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "movie id cannot be empty"})
			return
		}

		var movie Models.Movie

		err := moviecollection.FindOne(ctx, bson.M{"imdb_id": movieId}).Decode(&movie)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}

		c.JSON(http.StatusOK, movie)
	}
}

// ADD MOVIE

func Addmovies() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movie Models.Movie

		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "validation error",
				"detail": err.Error(),
			})
			return
		}

		result, err := moviecollection.InsertOne(ctx, movie)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// ADMIN REVIEW WITH AI SENTIMENT

func AdminReview() gin.HandlerFunc {
	return func(c *gin.Context) {

		movieId := c.Param("imdb_id")

		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
			return
		}

		var req struct {
			Adminreview string `json:"admin_review"`
		}

		var resp struct {
			RankingName string `json:"ranking_name"`
			Adminreview string `json:"admin_review"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		sentiment, rankVal, err := GetReview(req.Adminreview)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"imdb_id": movieId}

		update := bson.M{
			"$set": bson.M{
				"admin_review": req.Adminreview,
				"ranking": bson.M{
					"ranking_name":  sentiment,
					"ranking_value": rankVal,
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := moviecollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}

		resp.RankingName = sentiment
		resp.Adminreview = req.Adminreview

		c.JSON(http.StatusOK, resp)
	}
}

// AI SENTIMENT ANALYSIS

func GetReview(admin_review string) (string, int, error) {

	rankings, err := GetRankings()
	if err != nil {
		return "", 0, err
	}

	var sentiments []string

	for _, r := range rankings {
		if r.RankingValue != 999 {
			sentiments = append(sentiments, r.RankingName)
		}
	}

	sentimentDelimited := strings.Join(sentiments, ",")

	openAIKey := os.Getenv("OPENAI_APIKEY")

	if openAIKey == "" {
		return "", 0, errors.New("openai key not found")
	}

	llm, err := openai.New(openai.WithToken(openAIKey))
	if err != nil {
		return "", 0, err
	}

	basePrompt := os.Getenv("BASE_PROMPT")

	if basePrompt == "" {
		return "", 0, errors.New("base prompt missing")
	}

	prompt := strings.Replace(basePrompt, "{rankings}", sentimentDelimited, 1)

	response, err := llm.Call(context.Background(), prompt+admin_review)

	if err != nil {
		return "", 0, err
	}

	response = strings.TrimSpace(response)

	rankVal := 0

	for _, r := range rankings {
		if r.RankingName == response {
			rankVal = r.RankingValue
			break
		}
	}

	return response, rankVal, nil
}

// GET RANKINGS

func GetRankings() ([]Models.Ranking, error) {

	var rankings []Models.Ranking

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cursor, err := rankingscollection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}

	return rankings, nil
}

// RECOMMENDED MOVIES

func GetRecommendation() gin.HandlerFunc {
	return func(c *gin.Context) {

		userId, err := utilis.GetUserIdFromContext(c)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user id not found"})
			return
		}

		favoriteGenres, err := UserFavoriteGenres(userId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		limit := int64(5)

		limitStr := os.Getenv("RECOMMEND_MOVIE_LIMIT")

		if limitStr != "" {
			limit, _ = strconv.ParseInt(limitStr, 10, 64)
		}

		filter := bson.M{
			"genre.genre_name": bson.M{
				"$in": favoriteGenres,
			},
		}

		findOptions := options.Find()

		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value: -1}})
		findOptions.SetLimit(limit)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := moviecollection.Find(ctx, filter, findOptions)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer cursor.Close(ctx)

		var recommendedMovies []Models.Movie

		if err := cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, recommendedMovies)
	}
}

// USER FAVORITE GENRES

func UserFavoriteGenres(userId string) ([]string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userId}

	projection := bson.M{
		"favourite_genres.genre_name": 1,
		"_id":                         0,
	}

	opts := options.FindOne().SetProjection(projection)

	var results bson.M

	err := usercollection.FindOne(ctx, filter, opts).Decode(&results)

	if err != nil {

		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}

		return nil, err
	}

	favGenresArray, ok := results["favourite_genres"].(bson.A)

	if !ok {
		return []string{}, errors.New("unable to retrieve genres")
	}

	var genreNames []string

	for _, item := range favGenresArray {

		if genreMap, ok := item.(bson.M); ok {

			if name, ok := genreMap["genre_name"].(string); ok {

				genreNames = append(genreNames, name)

			}

		}

	}

	return genreNames, nil
}
