package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type CtxKey string

const connectionID CtxKey = "connectionID"

type WebsocketUpgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error)
}

func connect(wsu WebsocketUpgrader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsu.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		toSend := strings.Split("this is a backend message", " ")

		cID, ok := r.Context().Value(connectionID).(uuid.UUID)
		if !ok {
			http.Error(w, "some error", http.StatusInternalServerError)
			return
		}

		var i int
		for {
			send := struct {
				ID      uuid.UUID `json:"id,omitempty"`
				Message string    `json:"message,omitempty"`
			}{
				cID,
				toSend[i%len(toSend)],
			}

			if err := conn.WriteJSON(send); err != nil {
				// probably due to client close, break to trigger deferred close of conn
				break
			}

			time.Sleep(10 * time.Second)
			i++
		}
	}
}

func main() {

	liveMux := http.NewServeMux()
	liveMux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	go http.ListenAndServe(":9090", liveMux)

	mux := http.NewServeMux()

	// allow origins
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
		// org := r.Header.Get("Origin")
		// return org == "http://localhost:5173"
	}}

	mux.HandleFunc("/", connect(&upgrader))

	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")),
		Handler: cors.AllowAll().Handler(mux),
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, connectionID, uuid.New())
		},
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
