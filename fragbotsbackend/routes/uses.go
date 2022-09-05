package routes

import (
	"fragbotsbackend/constants"
	"fragbotsbackend/database"
	"fragbotsbackend/logging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

type Uses struct {
	Uses int `json:"uses"`
}

type PostUsesRequest struct {
	UUID string `json:"uuid" form:"uuid"`
}

func GetUses(c *gin.Context) {
	var uses Uses
	err := database.GetDocument("uses", bson.D{{"_id", "usesTracker"}}, &uses)
	if err != nil {
		logging.LogWarn("Failed to retrieve uses: " + err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong while retrieving uses"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"uses": uses.Uses})
}

func PostUses(c *gin.Context) {
	key := c.GetHeader("access-token")
	if key != constants.AccessToken {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Access Token"})
		return
	}
	var req PostUsesRequest
	err := c.Bind(&req)
	if err != nil || req.UUID == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Body"})
		return
	}
	err = database.UpdateDocumentIncField("users", bson.D{{"_id", req.UUID}}, bson.D{{"timesused", 1}})
	if err != nil {
		logging.LogWarn("Failed to add use to user, " + err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
		return
	}
	err = database.UpdateDocumentIncField("uses", bson.D{{"_id", "usesTracker"}}, bson.D{{"uses", 1}})
	if err != nil {
		logging.LogWarn("Failed to add use, " + err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to add use"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}
