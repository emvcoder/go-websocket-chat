package main

import (
	"fmt"
	"strconv"
	"net/http"
	"encoding/json"
	"github.com/gorilla/websocket"
)

type (
	NewMessage struct {
		key string
		message []byte
	}
	SimpleMessage struct {
		MessType, Message string
	}
)

var NewMessageChan = make(chan *NewMessage)

func ReadMessages(ws *websocket.Conn, r *http.Request) {
	for {
		_, messages, err := ws.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		NewMessageChan <- &NewMessage{r.RemoteAddr, messages}
	}
}

func decodingMessage(message *NewMessage) {
	var m SimpleMessage
	err := json.Unmarshal(message.message, &m)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	var user = GetUser(message.key)
	var currentRoom int = clients[message.key]

	switch(string(m.MessType)) {
	case "roomch":
		room, err := strconv.Atoi(m.Message)
		if err != nil {
			fmt.Println("Error: see 46 line*")
			return
		}

		if currentRoom != room {
			if len(Rooms[clients[user.key]].users) == 1 {
				CloseRoom(clients[user.key])
			} else {
				delete(Rooms[clients[user.key]].users, user.key)
			}

			go func() {
				Rooms[room].users[user.key] = user
				roomChan <- &RoomCh{user.key, room}
				WriteCommandAll(&wsCommand{nil, "cmd", Rooms})
				WriteMessage(&wsMessage{user.ws, user.key, "msg", SERVER_NAME, "You have successfully transferred from \""+strconv.Itoa(currentRoom)+"\" room to \""+strconv.Itoa(room)+"\" room"})
			}()
		} else {
			WriteMessage(&wsMessage{user.ws, user.key, "msg", SERVER_NAME, "You are already in this room."})
		}
	case "cmd":
		if len(m.Message) > 4 {
			WriteMessage(&wsMessage{user.ws, user.key, "msg", SERVER_NAME, "You changed your nickname from \""+user.name+"\" to \""+m.Message})
			user.name = m.Message
		} else {
			WriteMessage(&wsMessage{user.ws, user.key, "msg", SERVER_NAME, "Your nickname is less than 4 characters"})
		}
	case "message":
		if len(m.Message) > 0 {
			messagesChan <- &Message{message.key, user.name, m.Message}
		}
	}
}
