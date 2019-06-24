package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"golang.org/x/net/websocket"
)

// User is Client of Hub
type User struct {
	ID     int             `json:"id"`
	Name   string          `json:"name"`
	Output *websocket.Conn `json:"-"`
}

// Hub is the core of the server that handles User connections
type Hub struct {
	Users map[int]User
	Join  chan User
	Leave chan User
	Input chan Message
}

// Message is WS Message connection
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// Run is Hub's function than handles various cases of interaction with Users
func (hub *Hub) Run() {
	adminUser := User{ID: 0, Name: "ADMIN"}

	for {
		select {
		case user := <-hub.Join:
			hub.Users[user.ID] = user
			go func() {
				hub.Input <- Message{
					Type: "admin_message",
					Payload: map[string]interface{}{
						"text": fmt.Sprintf("%s joined", user.Name),
						"user": adminUser,
					},
				}

				err := websocket.JSON.Send(user.Output, user)
				if err != nil {
					fmt.Println("Error sending open data", err.Error())
				}
			}()
		case user := <-hub.Leave:
			delete(hub.Users, user.ID)
			go func() {
				hub.Input <- Message{
					Type: "admin_message",
					Payload: map[string]interface{}{
						"text": fmt.Sprintf("%s left", user.Name),
						"user": adminUser,
					},
				}
			}()
		case msg := <-hub.Input:
			for _, user := range hub.Users {
				err := websocket.JSON.Send(user.Output, msg)
				if err != nil {
					fmt.Println("Error broadcasting message: ", err.Error())
				}
			}
		}
	}
}

func main() {
	hub := &Hub{
		Users: make(map[int]User),
		Join:  make(chan User),
		Leave: make(chan User),
		Input: make(chan Message),
	}

	go hub.Run()
	http.Handle("/", websocket.Handler(func(conn *websocket.Conn) {
		user := &User{
			Output: conn,
			Name:   "",
			ID:     rand.Int(),
		}

		for {
			msg := Message{}
			err := websocket.JSON.Receive(conn, &msg)
			if err != nil {
				fmt.Println("Error receiving message", err.Error())
				hub.Leave <- *user
				return
			}
			switch msg.Type {
			case "login":
				user.Name = msg.Payload["name"].(string)
				hub.Join <- *user
				break
			case "message":
				msg := Message{
					Type: "message",
					Payload: map[string]interface{}{
						"user": *user,
						"text": msg.Payload["text"].(string),
					},
				}
				hub.Input <- msg
				break
			}
		}
	}))
	log.Println("Serving at http://localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
