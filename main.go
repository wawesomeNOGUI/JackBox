package main

import (
	//"flag"
	//"html/template"
	"log"
	"net/http"
	"encoding/json"
	"math/rand"
  "sync"
	"time"
	//"strings"

	"github.com/gorilla/websocket"
)

var adminMutex sync.Mutex //To make sure no concurrent writes to the same Conn occur
var tMutex sync.Mutex //T-Shirt Map Writes
var sMutex sync.Mutex //SnappyText Map Writes

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// users stores Conn and username
var users map[*websocket.Conn]string = make(map[*websocket.Conn]string)
var admin *websocket.Conn

//All the T-Shirt Drawings
var tshirts map[*websocket.Conn][]byte = make(map[*websocket.Conn][]byte)
//All The Snappy T-Shirt Text
var snappyText map[*websocket.Conn][]byte = make(map[*websocket.Conn][]byte)
//Map of Matching T-Shirt to Text (The peeps Conns who made text or shirt)
var match map[*websocket.Conn]*websocket.Conn = make(map[*websocket.Conn]*websocket.Conn)

//Scores when people vote on shirts
//var scores map[*websocket.Conn]int

var textTime bool = false

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
				//Doesn't need mutex because only admin can send "Start"
				k.WriteMessage(1, []byte("Start"))
			}
			continue
		}else if textTime {  //If not json then must be Snappy Text
			sMutex.Lock()
			log.Println(message)
			snappyText[c] = message           //single quotes to treat as rune == single char??
			sMutex.Unlock()
			//Voting Time and Display
			if len(snappyText) == len(users) {
				admin.WriteMessage(1, []byte("DisplayResults"))
				//Marshall the stuff
				//User Matches
				tempMap := make(map[string]string)
				for k, v := range match {
					tempMap[k.RemoteAddr().String()] = v.RemoteAddr().String()
				}

				message, err := json.Marshal(tempMap)
				if err != nil {
					log.Println(err)
				}
				admin.WriteMessage(1, message)

				//Shirts
				tempShirtMap := make(map[string]string)
				for k, v := range tshirts {
					tempShirtMap[k.RemoteAddr().String()] = string(v)
				}

				message, err = json.Marshal(tempShirtMap)
				if err != nil {
					log.Println(err)
				}
				admin.WriteMessage(1, message)

				//Text
				tempTextMap := make(map[string]string)
				for k, v := range snappyText {
					tempTextMap[k.RemoteAddr().String()] = string(v)
				}
				message, err = json.Marshal(tempTextMap)
				if err != nil {
					log.Println(err)
				}
				admin.WriteMessage(1, message)
			}
			continue
		}

		err = json.Unmarshal(message, &messageMap)
		if err != nil {
			log.Println("errorUnmarshal:", err)
			//Message Must Be T-Shirt Design Cause it's values don't map to map[string]string
			 tMutex.Lock()
			 tshirts[c] = message  //Just let the browsers JSON.parse the message
			 tMutex.Unlock()

			 //Doesn't need mutexes because only one GoRoutine will use this write section
			 if len(tshirts) == len(users) {
				 admin.WriteMessage(1, []byte("textSection"))
				 //Send Out T-Shirt Pics for users to make logos for
				 var tempSlice []*websocket.Conn
				 r1 := rand.New(rand.NewSource(time.Now().UnixNano()))

				 //Fill slice with users
				 for conn, _ := range users {
					 tempSlice = append(tempSlice, conn)
				 }

				 //Give each users a random shirt to make a slogan for
				 for k, _ := range tshirts {
					 x := r1.Intn(len(tempSlice))
					 for {
						 //Don't send this users shirt back to themselves
						 if tempSlice[x] != k {
							 break
						 }
						 x = r1.Intn(len(tempSlice))
					 }
					 //Write to match for the pair of peeps who made T-Shirt and Text
					 match[tempSlice[x]] = k
					 //Send user T-Shirt
					 k.WriteMessage(1, tshirts[tempSlice[x]])
				 }

				 textTime = true
			 }
		}

		if val, ok := messageMap["username"]; ok {
			users[c] = val
			//Add address to message so admin can make an object matching addr to name
			messageMap["addr"] = c.RemoteAddr().String()
			message, err = json.Marshal(messageMap)
			if err != nil {
				log.Println(err)
			}
			// 2 is binary message, 1 is text message
			adminMutex.Lock()
			admin.WriteMessage(1, message) //send admin usernames (won't ever lose names cause admin connects first to the server)
			adminMutex.Unlock()
		}

    //log.Println(string(message))
	}
}

func main() {
	fileServer := http.FileServer(http.Dir("./public"))
	http.HandleFunc("/echo", echo) //this request comes from webrtc.html
	http.Handle("/", fileServer)


	log.Fatal(http.ListenAndServe(":80", nil))
}
