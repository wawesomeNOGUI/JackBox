package main

import (
	//"flag"
	//"html/template"
	"log"
	"net/http"
	"encoding/json"
	//"fmt"
  //"sync"
	//"time"
	//"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// users stores Conn and username
var users map[*websocket.Conn]string = make(map[*websocket.Conn]string)
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
		if admin == nil {
			c.WriteMessage(1, []byte("Wait For Host To Connect First Then Reload This Page."))
			return
		}
		users[c] = "" //Add this connection to the list of users with a defualt blank username
		log.Println("user in")
	} else if string(message) == "admin" {
		admin = c
		log.Println("Admin in")
	} else {
		log.Println("Connection from non player")
		return
	}

	// ReadLoop
	messageMap := make(map[string]string)

	for {
		_, message, err := c.ReadMessage() //ReadMessage blocks until SDP message received
		if err != nil {
			log.Println("read:", err)
			if c == admin {
				admin = nil
			} else {
				delete(users, c)
				disconnectMessage := make(map[string]string)
				disconnectMessage["disconnect"] = c.RemoteAddr().String()
				temp, err := json.Marshal(disconnectMessage)
				if err != nil {
					log.Println(err)
				}
				//Tell admin which user disconnected
				admin.WriteMessage(1, temp)
			}
			break
		}

		if string(message) == "Start" {
			for k, _ := range users {
				k.WriteMessage(1, []byte("Start"))
			}
			continue
		}

		err = json.Unmarshal(message, &messageMap)
		if err != nil {
			log.Println("errorUnmarshal:", err)
		}

		if val, ok := messageMap["username"]; ok {
			users[c] = val
			//Add address to message
			messageMap["addr"] = c.RemoteAddr().String()
			message, err = json.Marshal(messageMap)
			if err != nil {
				log.Println(err)
			}
			// 2 is binary message, 1 is text message
			admin.WriteMessage(1, message) //send admin usernames (won't ever lose names cause admin connects first to the server)
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
