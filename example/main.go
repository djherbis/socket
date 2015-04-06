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

	server.On(socket.Connection, func(so socket.Socket) {
		so.Emit("hello", "world")
		so.On("hey", func(msg string) {
			fmt.Println("WORKING")
			fmt.Println(msg)
		})
	})

	server.Of("/views").On(socket.Connection, func(so socket.Socket) {
		so.Join("group")
		fmt.Println(so.ID(), "joined")
		so.To("group").Emit("hello", so.ID())
		so.On("hey", func(msg string) {
			fmt.Println(so.ID(), msg)
		})
		so.On(socket.Disconnect, func() {
			fmt.Println(so.ID(), "leaving")
		})
	})

	server.Of("/views").On(socket.Disconnection, func(so socket.Socket) {
		fmt.Println(so.ID(), "left")
		so.To("group").Emit("goodbye", so.ID())
	})

	go func() {
		if so, err := socket.New("localhost/views"); err == nil {
			so.Emit("hey", ":)")
			so.On("hello", func(msg string) {
				fmt.Println(msg, ": hello")
			})
		} else {
			fmt.Println(err.Error())
		}
	}()

	router := http.NewServeMux()
	router.Handle("/socket", server)
	router.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe("localhost:80", router)
}
