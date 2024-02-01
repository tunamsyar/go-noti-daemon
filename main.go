package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/takama/daemon"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type MessageRequest struct {
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	DeviceTokens []string `json:"device_tokens"`
}

var server *http.Server

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {

	usage := "Usage: matt-daemon install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			http.HandleFunc("/send_notifications", sendNotifications)
			log.Fatal(http.ListenAndServe(":8080", nil))
			return service.Start()
		case "stop":
			log.Println("Received request to stop the server. Shutting down gracefully...")
			shutdown()
			log.Println("Server shut down successfully")
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	return usage, nil
}

func main() {
	//
	service, err := daemon.New("matt-daemon", "The demon formerly known as Matt Damon", daemon.UserAgent)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	daemonService := &Service{service}
	status, err := daemonService.Manage()

	if err != nil {
		log.Fatal(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}

func sendNotifications(w http.ResponseWriter, r *http.Request) {
	var request MessageRequest

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	opt := option.WithCredentialsFile("auth.json")
	app, err := firebase.NewApp(ctx, nil, opt)

	if err != nil {
		http.Error(w, "Firebase initialization failed", http.StatusInternalServerError)
		return
	}

	fcmClient, err := app.Messaging(ctx)

	if err != nil {
		http.Error(w, "FCM Client init failed", http.StatusInternalServerError)
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: request.Title,
			Body:  request.Body,
		},
		Token: request.DeviceTokens[0], // Assuming a single device token for simplicity. Still janky. Should be a loop.
	}

	response, err := fcmClient.Send(ctx, message)
	if err != nil {
		http.Error(w, "Failed to send FCM message", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully sent message to device: %v", response)
}

func shutdown() {
	if server != nil {
		server.Shutdown(context.Background())
	}
	os.Exit(0)
}
