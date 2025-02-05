package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]Client)
var broadcast = make(chan MessageWithCoords)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	MessageTypeAuth      = 0
	MessageTypeMouseMove = 1
)

type Coords struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MessageWithCoords struct {
	SenderId    string `json:"senderId"`
	MessageType int    `json:"messageType"`
	Message     struct {
		Coords Coords `json:"coords"`
	} `json:"message"`
}

type MessageWithClients struct {
	SenderId    string `json:"senderId"`
	MessageType int    `json:"messageType"`
	Message     struct {
		Clients []Client `json:"clients"`
	} `json:"message"`
}

type Client struct {
	Id     string `json:"id"`
	Coords Coords `json:"coords"`
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Ошибка сервера:", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка подключения: %v", err)
		return
	}
	defer conn.Close()

	fmt.Println("Новое подключение:", conn.RemoteAddr())

	for {
		var msg MessageWithCoords

		err := conn.ReadJSON(&msg)

		if err != nil {
			log.Printf("Ошибка чтения: %v", err)
			delete(clients, conn)
			break
		}

		addClient(msg, conn)

		if msg.MessageType == MessageTypeAuth {
			var clientsList []Client

			for _, client := range clients {
				clientsList = append(clientsList, client)
			}

			var clientsMessage = MessageWithClients{
				SenderId:    msg.SenderId,
				MessageType: msg.MessageType,
				Message: struct {
					Clients []Client `json:"clients"`
				}{
					Clients: clientsList,
				},
			}

			conn.WriteJSON(clientsMessage)
		} else {
			broadcast <- msg
		}
		fmt.Println(clients)
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {

			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Ошибка отправки: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func addClient(msg MessageWithCoords, conn *websocket.Conn) {
	clients[conn] = Client{Id: msg.SenderId, Coords: msg.Message.Coords}
}
