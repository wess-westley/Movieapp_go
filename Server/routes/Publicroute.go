package routes

import (
	controllers "Magic/controllers"

	"github.com/gin-gonic/gin"
)

func SetPublicRoute(router *gin.Engine) {
	router.GET("/Movies", controllers.GetMovies())

	router.POST("/registeruser", controllers.RegisterUser())
	router.POST("/Login", controllers.UserLogin())

}
