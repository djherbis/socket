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

	server.Of("/views").On(socket.CONNECTION, func(so *socket.Socket) {
		so.Join("group")
		so.To("group").Emit("hello", so.Id())
		so.On("echo", func(msg string) {
			fmt.Println("echo")
			so.To("group").Emit("hello", msg)
		})
	})

	server.Of("/views").On(socket.DISCONNECTION, func(so *socket.Socket) {
		fmt.Println("DISCONNECTED")
	})

	router := http.NewServeMux()
	router.Handle("/socket", server)
	router.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe("localhost:80", router)
}
