package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 1024
	pongWait       = 60 * time.Second
	pingWait       = (pongWait * 9) / 10
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade HTTP Connection %v\n", err)

			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Not Found")
			return
		}

		defer ws.Close()

		ws.SetReadLimit(maxMessageSize)
		ws.SetReadDeadline(time.Now().Add(pongWait))
		ws.SetWriteDeadline(time.Now().Add(pingWait))
		ws.SetPongHandler(func(appData string) error {
			log.Println("Received Pong. ", appData)
			ws.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
		ws.SetPingHandler(func(appData string) error {
			log.Println("Received Ping. ", appData)
			return nil
		})
		ws.SetCloseHandler(func(code int, text string) error {
			log.Println("Closing conn. ", code, text)
			return nil
		})

		for {
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("Error reading message:", err)
				}
				break
			}

			fmt.Printf("Received: %s\n", msg)

			if err := ws.WriteMessage(msgType, msg); err != nil {
				log.Println("Error write message:", err)
				break
			}

			w, err := ws.NextWriter(websocket.PingMessage)
			if err != nil {
				return
			}

			w.Write([]byte{' '})

			if err := w.Close(); err != nil {
				return
			}
		}
	})

	log.Println("Server started on :8008")
	if err := http.ListenAndServe(":8008", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
