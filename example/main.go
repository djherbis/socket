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

	server.Of("/views").On(socket.Connection, func(so socket.Socket) {
		so.Join("group")
		so.To("group").Emit("hello", so.Id())
		so.On("echo", func(msg string) {
			fmt.Println("echo")
			so.To("group").Emit("hello", msg)
		})
		so.On("disconnect", func() {
			fmt.Println("hello")
		})
	})

	server.Of("/views").On(socket.Disconnection, func(so socket.Socket) {
		so.To("group").Emit("hello", "goodbye!")
	})

	router := http.NewServeMux()
	router.Handle("/socket", server)
	router.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe("localhost:80", router)
}
