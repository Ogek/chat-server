package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"golang.org/x/net/websocket"
)

type Message struct {
	Text string `json:"text"`
	User User   `json:"user"`
}

type User struct {
	Id     int
	Output *websocket.Conn `json:"-"`
}

type ChatServer struct {
	Users map[int]User
	Join  chan User
	Leave chan User
	Input chan Message
}

func (cs *ChatServer) Run() {
	for {
		select {
		case user := <-cs.Join:
			cs.Users[user.Id] = user
			go func() {
				cs.Input <- Message{
					Text: fmt.Sprintf("%d joined", user.Id),
					User: User{Id: 0},
				}
			}()
		case user := <-cs.Leave:
			delete(cs.Users, user.Id)
			go func() {
				cs.Input <- Message{
					Text: fmt.Sprintf("%d left", user.Id),
					User: User{Id: 0},
				}
			}()
		case msg := <-cs.Input:
			for _, user := range cs.Users {
				err := websocket.JSON.Send(user.Output, msg)
				if err != nil {
					fmt.Println("Error broadcasting message: ", err.Error())
				}
			}
		}
	}
}

func main() {
	chatServer := &ChatServer{
		Users: make(map[int]User),
		Join:  make(chan User),
		Leave: make(chan User),
		Input: make(chan Message),
	}

	go chatServer.Run()
	http.Handle("/", websocket.Handler(func(ws *websocket.Conn) {
		user := User{
			Output: ws,
			Id:     rand.Int(),
		}
		chatServer.Join <- user

		for {
			msg := Message{
				User: user,
			}
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				chatServer.Leave <- user
				return
			}
			chatServer.Input <- msg
		}

	}))
	log.Println("Serving at http://localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
