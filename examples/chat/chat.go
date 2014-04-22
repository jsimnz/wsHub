// A simple chat program leveraging websockets and the wsHub package
package main

import (
	"fmt"
	"net/http"

	"../../../wsHub"
)

var (
	hub *wsHub.WsHub
)

func main() {
	fmt.Println("Creating WebSocket comm hub..")
	hub = wsHub.NewHub()
	fmt.Println("Running hub...")
	go hub.Run()

	http.HandleFunc("/ws/chat", wsHandler)
	fmt.Println("Starting web server...")
	http.ListenAndServe(":9090", nil)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got new connection")
	client, err := wsHub.NewClient(w, r)
	if err != nil {
		panic(err)
	}

	hub.RegisterClient(client)

	go client.Start()
	defer func() {
		fmt.Println("Removing client from hub")
		hub.UnregisterClient(client)
	}()

	for {
		resp, err := client.Read()
		if err != nil {
			break
		}
		hub.Broadcast(client, resp)
	}
}
