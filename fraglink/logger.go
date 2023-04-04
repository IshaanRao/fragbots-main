package main

import (
	"log"
)

var colorReset = "\033[0m"

var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"
var colorCyan = "\033[36m"

func LogWarn(message string) {
	println(colorYellow + "[FragLink-WARN] " + message + colorReset)
}
func Log(message string) {
	println(colorGreen + "[FragLink] " + message + colorReset)
}

func LogFatal(message string) {
	log.Fatal(colorRed + "[FragLink-FATAL] " + message + colorReset)
}
