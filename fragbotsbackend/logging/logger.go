package logging

import (
	"log"
)

var colorReset = "\033[0m"

var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"
var colorCyan = "\033[36m"

func Debug(message string) {
	println(colorCyan + "[FragBotBackend-DEBUG] " + message + colorReset)
}

func LogWarn(message string) {
	println(colorYellow + "[FragBotBackend-WARN] " + message + colorReset)
}
func Log(message string) {
	println(colorGreen + "[FragBotBackend] " + message + colorReset)
}

func LogFatal(message string) {
	log.Fatal(colorRed + "[FragBotBackend-FATAL] " + message + colorReset)
}
