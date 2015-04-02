package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/djherbis/socket"
)

func main() {
	server := socket.NewServer()

	server.On(socket.Connection, func(so socket.Socket) {
		log.Printf("localhost/hook/%s", so.Id())
	})

	server.On(socket.Disconnection, func(so socket.Socket) {
		log.Println("left", so.Id())
	})

	router := http.NewServeMux()
	router.HandleFunc("/hook/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/hook/")
		server.To(id).Emit("hook", "say hello :)")
	})

	router.Handle("/socket", server)
	http.ListenAndServe("localhost:80", router)
}
