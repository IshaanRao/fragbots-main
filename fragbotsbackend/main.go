package main

import (
	"context"
	"errors"
	"fragbotsbackend/constants"
	"fragbotsbackend/database"
	"fragbotsbackend/logging"
	"fragbotsbackend/routes"
	gin "github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var Router *gin.Engine
var srv *http.Server
var ReqClient = req.C().
	SetTimeout(20 * time.Second)

func init() {
	Router = gin.Default()
}
func main() {
	constants.LoadConsts()
	database.StartClient()
	routes.Router = Router
	routes.InitRoutes()

	srv = &http.Server{
		Addr:    ":" + strconv.FormatInt(int64(constants.Port), 10),
		Handler: Router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()

	gracefulStop()
}

func gracefulStop() {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Debug("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logging.LogFatal("Server forced to shutdown: " + err.Error())
	}
	if err := database.MongoClient.Disconnect(context.TODO()); err != nil {
		logging.LogWarn("Failed to disconnect mongo successfully")
	}
	logging.Debug("Successfully closed mongo client")

	logging.Log("Server exiting")
}
