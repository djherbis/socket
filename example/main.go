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
		fmt.Println(so.Id(), "joined")
		so.To("group").Emit("hello", so.Id())
		so.On("hey", func(msg string) {
			fmt.Println(so.Id(), msg)
		})
		so.On(socket.Disconnect, func() {
			fmt.Println(so.Id(), "leaving")
		})
	})

	server.Of("/views").On(socket.Disconnection, func(so socket.Socket) {
		fmt.Println(so.Id(), "left")
		so.To("group").Emit("goodbye", so.Id())
	})

	router := http.NewServeMux()
	router.Handle("/socket", server)
	router.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe("localhost:80", router)
}
