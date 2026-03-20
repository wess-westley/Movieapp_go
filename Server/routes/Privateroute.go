package routes

import (
	middleware "Magic/Middleware"
	controllers "Magic/controllers"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetProtectedRoute(router *gin.Engine, client *mongo.Client) {
	router.Use(middleware.AuthMiddleWare())

	router.GET("/movie/:imdb_id", controllers.GetMovie(client))
	router.POST("/addmovie", controllers.AddMovie(client))
	router.GET("/recommendedmovies", controllers.GetRecommendedMovies(client))
	router.PATCH("/updatereview/:imdb_id", controllers.AdminReviewUpdate(client))

}
