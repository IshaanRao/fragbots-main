package main

import (
	"log"
	"net/http"
)

var addr = "localhost:2468"

func main() {
	http.HandleFunc("/ws", wsHandler)
	Log("Started websocket server at: " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
