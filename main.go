package main

import (
	//"flag"
	//"html/template"
	"log"
	"net/http"
	//"fmt"
//	"sync"
	//"time"
	//"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

//var users []websocket.Conn

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage() //ReadMessage blocks until SDP message received
		if err != nil {
			log.Println("read:", err)
		}

    log.Println(message)

	}
}


func main() {
	fileServer := http.FileServer(http.Dir("./public"))
	http.HandleFunc("/echo", echo) //this request comes from webrtc.html
	http.Handle("/", fileServer)


	log.Fatal(http.ListenAndServe(":80", nil))

}
