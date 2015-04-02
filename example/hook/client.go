package main

import (
	"fmt"
	"log"

	"github.com/djherbis/socket"
)

func main() {

	so, err := socket.New("localhost")
	if err != nil {
		log.Fatal(err.Error())
	}

	wait := make(chan struct{})

	so.On("hook", func(msg string) {
		fmt.Println(msg)
		wait <- struct{}{}
	})

	so.On(socket.Disconnect, func() {
		log.Println("Server hung-up")
		wait <- struct{}{}
	})

	<-wait
}
