package routes

import (
	"fragbotsbackend/constants"
	"github.com/gin-gonic/gin"
	"net/http"
)

var Router *gin.Engine

func InitRoutes() {
	// BotInfo Routes
	botInfo := Router.Group("/bots")
	botInfo.Use(auth())
	botInfo.GET("/:botid", getBotData)
	botInfo.POST("/:botid", PostBot)
}

func auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		key := context.GetHeader("access-token")
		if key != constants.AccessToken {
			context.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid Access Token"})
		} else {
			context.Next()
		}
	}
}
