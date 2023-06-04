package logging

import (
	"fmt"
	"log"
	"os"
)

// Color coded to make logs easier to read
var colorReset = "\033[0m"
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"

func LogWarn(v ...any) {
	log.Println(colorYellow+"[WARN]", fmt.Sprintln(v...), colorReset)
}

func Log(v ...any) {
	log.Println(colorGreen+"[INFO]", fmt.Sprintln(v...), colorReset)
}

func LogFatal(v ...any) {
	log.Println(colorRed+"[FATAL]", fmt.Sprintln(v...), colorReset)
	os.Exit(1)
}
