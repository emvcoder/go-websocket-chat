package main

import (
	"fmt"
	// "strconv"
	"net/http"
	"io/ioutil"
	"html/template"
	"github.com/gorilla/websocket"
)

type (
	Client struct {
		ws *websocket.Conn
		key, name string
	}
	Room struct {
		Id int
		name string
		users map[string]*Client
	}
	Message struct {
		key, Author, Message string
	}
	wsMessage struct {
		ws *websocket.Conn
		key, MessType, Author, Message string
	}
	wsCommand struct {
		ws *websocket.Conn
		MessType string
		Rooms map[int]*Room
	}
	RoomCh struct {
		key string
		room int
	}
)

const (
	SERVER_NAME string = "System" // the author of system messages
	ROOM_CAPACITY int = 2 // maximum capacity of rooms
)

var (
	roomCol int
	Rooms = make(map[int]*Room)
	clients = make(map[string]int)
	roomChan = make(chan *RoomCh)
	messagesChan = make(chan *Message, 100)
	messagesQueue = make(chan *wsMessage, 100)
)

// Creates new room
func CreateRoom(client *Client, key string) {
	var soclient = make(map[string]*Client)
	roomCol++
	soclient[key] = client
	Rooms[roomCol] = &Room{roomCol, string(roomCol), soclient}
	roomChan <- &RoomCh{client.key, roomCol}
}

func GetUser(key string) *Client {
	return Rooms[clients[key]].users[key]
}

func CloseRoom(key int) {
	delete(Rooms, key)
}

// Distributes the recipients to the rooms
func Distributor(ws *websocket.Conn, address string) {
	var handled bool = false
	for key, main := range Rooms {
		if len(main.users) < ROOM_CAPACITY {
			main.users[address] = &Client{ws, address, address}
			roomChan <- &RoomCh{address, key}
			handled = true
		}

		if (handled == true) && (len(main.users) == 0) {
			CloseRoom(key)
		}
	}

	if handled != true {
		CreateRoom(&Client{ws, address, address}, address)
	}

	go func() {
		WriteCommand(&wsCommand{ws, "cmd", Rooms})
		WriteMessage(&wsMessage{ws, address, "msg", SERVER_NAME, "Welcome"})
		for msg := range messagesQueue {
			WriteMessage(msg)
		}
	}()
}

func wsHandler(w http.ResponseWriter, r *http.Request)  {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)

	if err != nil {
		fmt.Println(err)
	}

	Distributor(ws, r.RemoteAddr)
	defer func() {
		ws.Close()
		if len(Rooms[clients[r.RemoteAddr]].users) == 1 {
			CloseRoom(clients[r.RemoteAddr])
		} else {
			delete(Rooms[clients[r.RemoteAddr]].users, r.RemoteAddr)
		}
		delete(clients, r.RemoteAddr)
	}()

	ReadMessages(ws, r)
}

// Router
func router() {
	for {
		select {
		case message := <- NewMessageChan:
			decodingMessage(message)
		case data := <- roomChan:
			clients[data.key] = data.room
		case message := <- messagesChan:
			fmt.Println("Message from", "\""+message.Author+"\":", message.Message)
			for _, main := range Rooms[clients[message.key]].users {
				messagesQueue <- &wsMessage{main.ws, message.key, "msg", message.Author, message.Message}
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	var homeTempl = template.Must(template.ParseFiles("templates/home.html"))
	homeTempl.Execute(w, nil)
}

func readFile(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadFile("."+r.URL.Path)
	w.Header().Set("Content-type", "text/css")
	w.Write([]byte(body));
}

func main() {
	fmt.Println("Server started on port 8080")

	go router()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			handler(w, r)
		case "/ws":
			wsHandler(w, r)
		case LookForMatch(r.URL.Path):
			readFile(w, r)
		}
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error when tryed to create http server")
	}
}
