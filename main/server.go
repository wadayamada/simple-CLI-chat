package main

import (
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		log.Println("connected: ", s.ID())
		return nil
	})

	server.OnEvent("/", "chat message", func(s socketio.Conn, usr string, msg string) string {
		data := msg + " from " + usr;
		server.BroadcastToNamespace("/", "broadcast", data);
		return "recv " + msg
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}