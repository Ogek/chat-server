package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

// User is Client of Hub
type User struct {
	ID     int             `json:"id"`
	Name   string          `json:"name"`
	Output *websocket.Conn `json:"-"`
}

// Message is WS Message connection
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// Hub is the core of the server that handles User connections
type Hub struct {
	Users map[int]User
	Join  chan User
	Leave chan User
	Input chan Message
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		Users: make(map[int]User),
		Join:  make(chan User),
		Leave: make(chan User),
		Input: make(chan Message),
	}
}

// Run runs the chat hub, handling websocket connections,
// as well as various cases of user interaction
func (hub *Hub) Run() {
	go func() {
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
	}()

	http.Handle("/", websocket.Handler(hub.handleWS))
}

// handleWS handles a wesocket connection to client
func (hub *Hub) handleWS(conn *websocket.Conn) {
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
}

func main() {
	hub := NewHub()
	hub.Run()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Serving at http://localhost:%s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
