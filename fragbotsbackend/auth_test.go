package main

import (
	"fragbotsbackend/constants"
	"fragbotsbackend/database"
	"fragbotsbackend/routes"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth(t *testing.T) {
	r := preTest()
	r.GET("/:botid", routes.GetBot)
	req, _ := http.NewRequest("GET", "/Verified1", nil)
	req.Header.Set("access-token", constants.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	responseData, _ := io.ReadAll(w.Body)
	print(responseData)
}

func preTest() *gin.Engine {
	constants.LoadConsts()
	database.StartClient()
	return SetUpRouter()
}
func SetUpRouter() *gin.Engine {
	router := gin.Default()
	return router
}
