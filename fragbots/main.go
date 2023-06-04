package main

import (
	"github.com/Prince/fragbots/api"
	"github.com/Prince/fragbots/logging"
)

func main() {
	err := api.StartApi()
	if err != nil {
		logging.LogFatal("Shutdown: ", err)
	}
	logging.LogFatal("Server somehow shutdown with no error so something went terribly wrong")
}
