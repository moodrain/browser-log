package main

import (
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
)

var msgQueue = make(chan string, 10)
var conn *websocket.Conn

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrade := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	oldConn := conn
	newConn, _ := upgrade.Upgrade(w, r, nil)
	conn = newConn
	if oldConn != nil {
		_ = oldConn.Close()
	}
}

func HttpHandler(w http.ResponseWriter, r *http.Request) {
	if conn != nil && r.FormValue("log") != "" || r.Method == "POST" {
		select {
		case msgQueue <- r.FormValue("log"):
		default:
		}
	} else {
		page, _ := ioutil.ReadFile("index.html")
		_, _ = w.Write(page)
	}
}

func main() {
	go func() {
		for {
			select {
			case msg := <-msgQueue:
				{
					if conn != nil {
						err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
						if err != nil {
							_ = conn.Close()
							conn = nil
						}
					}
				}
			}
		}
	}()
	http.HandleFunc("/ws", WebsocketHandler)
	http.HandleFunc("/", HttpHandler)
	_ = http.ListenAndServe("127.0.0.1:8080", nil)
}
