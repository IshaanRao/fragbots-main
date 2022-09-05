package routes

import (
	"fragbotsbackend/database"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

type FragBotsUser struct {
	Id          string `json:"_id" bson:"_id"`
	TimesUsed   int    `json:"timesused" bson:"timesused"`
	Discord     string `json:"discord" bson:"discord"`
	Blacklisted bool   `json:"blacklisted" bson:"blacklisted"`
	Whitelisted bool   `json:"whitelisted" bson:"whitelisted"`
	Exclusive   bool   `json:"exclusive" bson:"exclusive"`
	Active      bool   `json:"active" bson:"active"`
	Priority    bool   `json:"priority,omitempty" bson:"priority,omitempty"`
}

func GetUser(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Missing UUID"})
		return
	}

	var user FragBotsUser
	err := database.GetDocument("users", bson.D{{"_id", uuid}}, &user)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}
	c.IndentedJSON(http.StatusOK, user)
}
