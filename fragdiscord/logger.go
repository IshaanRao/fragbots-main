package main

import (
	"log"
)

var colorReset = "\033[0m"

var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"
var colorCyan = "\033[36m"

func Debug(message string) {
	if DebugMode {
		println(colorCyan + "[" + Name + "-DEBUG] " + message + colorReset)
	}
}

func LogWarn(message string) {
	println(colorYellow + "[" + Name + "-WARN] " + message + colorReset)
}
func Log(message string) {
	println(colorGreen + "[" + Name + "] " + message + colorReset)
}

func LogFatal(message string) {
	log.Fatal(colorRed + "[" + Name + "-FATAL] " + message + colorReset)
}
