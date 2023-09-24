package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

func main() {

	mux := http.NewServeMux()

	// allow origins
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true

		// org := r.Header.Get("Origin")
		// return org == "http://localhost:5173"
	}}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		var i int
		for {
			send := struct {
				Message int `json:"message,omitempty"`
			}{i}

			if err := conn.WriteJSON(send); err != nil {
				// probably due to client close, break to trigger deferred close of conn
				break
			}

			time.Sleep(time.Second / 4)
			i++
		}
	})

	http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")), mux)
}
