package main

import (
	// "fmt"
	"encoding/json"
	"github.com/gorilla/websocket"
)

func WriteMessage(message *wsMessage) {
	b, _ := json.Marshal(message)
	message.ws.WriteMessage(websocket.TextMessage, b)
}

func WriteCommand(command *wsCommand) {
	b, _ := json.Marshal(command)
	command.ws.WriteMessage(websocket.TextMessage, b)
}

func WriteCommandAll(command *wsCommand) {
	for key, room := range clients {
		user := Rooms[room].users[key]
		b, _ := json.Marshal(command)
		user.ws.WriteMessage(websocket.TextMessage, b)
	}
}