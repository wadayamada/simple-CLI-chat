package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	socketio "github.com/googollee/go-socket.io"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:example@localhost:27017"))

	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		log.Println("connected: ", s.ID())
		return nil
	})

	server.OnEvent("/", "archive", func(s socketio.Conn, usr string) {
		collection := client.Database("chat").Collection(usr)
		cursor, err := collection.Find(context.TODO(), bson.D{})
		var results []bson.M
		if err = cursor.All(context.TODO(), &results); err != nil {
			log.Fatal(err)
		}
		var slice []string
		for _, result := range results {
			slice = append(slice, result["content"].(string))
		}
		s.Emit("archive", strings.Join(slice,","))
	})

	server.OnEvent("/", "chat message", func(s socketio.Conn, usr string, msg string) string {
		data := msg + " from " + usr;
		server.BroadcastToNamespace("/", "broadcast", data);
		collection := client.Database("chat").Collection(usr)
		if collection == nil {
			var db *mongo.Database
			err := db.CreateCollection(context.TODO(), usr)
			if err != nil {
				log.Fatal(err)
			}
			collection := client.Database("chat").Collection(usr)
			_, err = collection.InsertOne(context.TODO(), bson.D{{"content", msg}})
		} else{
			_, _ = collection.InsertOne(context.TODO(), bson.D{{"content", msg}})
		}
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