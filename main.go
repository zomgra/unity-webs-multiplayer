package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	id   string
	conn *websocket.Conn
	x, y float64
}

var (
	clients      = make(map[string]*Client)
	clientsMutex sync.Mutex
	upgrader     = websocket.Upgrader{}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	var msg Message
	err = conn.ReadJSON(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = conn.WriteJSON(&msg)
	client := &Client{
		id:   msg.Id, // Используем Id из сообщения
		conn: conn,
		x:    msg.X,
		y:    msg.Y,
	}

	clientsMutex.Lock()
	clients[client.id] = client
	clientsMutex.Unlock()

	broadcastAllClientInfo()

	// notifyNewUserId(client)
	notifyNewUser(client)

	for {
		//var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			clientsMutex.Lock()
			delete(clients, msg.Id)
			clientsMutex.Unlock()
			fmt.Println(err)
			break
		}
		log.Println(msg)

		clientsMutex.Lock()
		client.x = msg.X
		client.y = msg.Y
		clientsMutex.Unlock()
		// Рассылаем обновленные координаты всем подключенным клиентам
		broadcastAllClientInfo()
	}
}

func notifyNewUserId(c *Client) {

}

func notifyNewUser(newClient *Client) {
	msg := Message{
		Id: newClient.id,
		X:  newClient.x,
		Y:  newClient.y,
	}

	// Преобразование в JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}

	// Отправка JSON через WebSocket всем, кроме нового пользователя
	for _, client := range clients {

		err := client.conn.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Println("Error sending message to client:", err)
		}

	}
}

func sendClientInfo(client *Client) {
	msg := Message{
		Id: client.id,
		X:  client.x,
		Y:  client.y,
	}

	// Преобразование в JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}

	// Отправка JSON через WebSocket
	err = client.conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Println("Error sending message to client:", err)
	}
}

func broadcastAllClientInfo() {
	for _, client := range clients {
		sendClientInfo(client)
		log.Println("Sent message to client", client.id)
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections) // Изменил путь на /ws
	fmt.Println("Server is running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

type Message struct {
	Id string  `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

func generateID() string {
	return fmt.Sprint(rand.Intn(1000))
}
