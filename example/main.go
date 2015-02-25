package main

import (
	"fmt"
	"net/http"

	"github.com/djherbis/socket"
)

type Message struct {
	Test string `json:"test"`
}

func main() {
	server := socket.NewServer()

	server.On("connection", func(so *socket.Socket) {
		so.Emit("test", 1, 2, 3)
		so.Emit("test", 3, 3, 3)

		so.On("test", func(msg *Message) {
			fmt.Println(msg.Test)
		})
	})

	server.On("disconnection", func(so *socket.Socket) {
		fmt.Println("broken")
	})

	router := http.NewServeMux()
	router.Handle("/socket", server)
	router.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe("localhost:80", router)
}
