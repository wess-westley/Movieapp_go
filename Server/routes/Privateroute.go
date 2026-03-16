package routes

import (
	middleware "Magic/Middleware"
	controllers "Magic/controllers"

	"github.com/gin-gonic/gin"
)

func SetProtectedRoute(router *gin.Engine) {
	router.Use(middleware.Authmiddleware())
	router.GET("/Movie/:imdb_id", controllers.GetMovie())
	router.POST("/Add", controllers.Addmovies())
	router.GET("/recommendedmovies", controllers.GetRecommendation())

}
