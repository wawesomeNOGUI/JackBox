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

var users map[*websocket.Conn]int = make(map[*websocket.Conn]int)
var admin *websocket.Conn

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// First message should be if the client is a user or the admin
	_, message, err := c.ReadMessage() //ReadMessage blocks until SDP message received
	if err != nil {
		log.Println("read:", err)
	}

	if string(message) == "user" {
		users[c] = 1 //Add this connection to the list of users
		log.Println("user in")
	} else if string(message) == "admin" {
		admin = c
		log.Println("Admin in")
	} else {
		log.Println("Connection from non player")
		return
	}

	for {
		_, message, err := c.ReadMessage() //ReadMessage blocks until SDP message received
		if err != nil {
			log.Println("read:", err)
			if c == admin {
				admin = nil
			} else {
				delete(users, c)
			}
			break
		}

    log.Println(string(message))

	}
}


func main() {
	fileServer := http.FileServer(http.Dir("./public"))
	http.HandleFunc("/echo", echo) //this request comes from webrtc.html
	http.Handle("/", fileServer)


	log.Fatal(http.ListenAndServe(":80", nil))

}
