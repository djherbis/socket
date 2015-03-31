socket 
==========

[![GoDoc](https://godoc.org/github.com/djherbis/socket?status.svg)](https://godoc.org/github.com/djherbis/socket)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.txt)

Simple Socket.io alternative with #golang server

Usage
------------

Server:

```go
package main

import (
  "fmt"
  "net/http"

  "github.com/djherbis/socket"
)

type MyObject struct{
  Text string `json:"text"`
}

func main() {
  server := socket.NewServer()

  server.On(socket.Connection, func(so socket.Socket) {
    so.Emit("hello", "world")

    so.On("hey", func(msg string) {
      fmt.Println(msg)
    })

    so.On("new obj", func(myobj MyObject){
      fmt.Println(myobj)
    })
  })

  router := http.NewServeMux()
  router.Handle("/socket", server)
  router.Handle("/", http.FileServer(http.Dir("."))) // serve up socket.js
  http.ListenAndServe("localhost:8080", router)
}
```

Client:

```html
<script src="socket.js"></script>
<script>
  var socket = io("localhost:8080/");
  socket.emit("hey", "hey there!");
  socket.on("hello", function(msg){
    console.log(msg);
  });
  socket.emit("new obj", {text: "objects are automatically marshalled/unmarshalled"});
</script>
```

Installation
------------
```sh
# Server Side
go get github.com/djherbis/socket

# Client Side
<script src="socket.js"></script>
```
