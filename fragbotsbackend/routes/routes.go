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
	botInfo.GET("/:botid", GetBot)
	botInfo.POST("/:botid", PostBot)
	botInfo.DELETE("/:botid", DeleteBot)
	botInfo.PUT("/:botid", PutBot)
	Router.GET("/users/:uuid", GetUser)
	Router.GET("/uses", GetUses)
	Router.POST("/uses", PostUses)

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
