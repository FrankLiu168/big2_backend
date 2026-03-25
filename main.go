package main

import (
	"fmt"
	"log"
	"net/http"
	"big2backend/connector"
	"big2backend/api"
)

func main() {
	startAPIServer()
}

func startAPIServer() {
	api.StartAPI()
}

func startSocketServer() {
	http.HandleFunc("/ws", connector.HandleWebSocket)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}