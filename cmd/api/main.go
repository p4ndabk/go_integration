package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/joho/godotenv"
)

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

var topic *pubsub.Topic

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Não foi possível carregar .env")
	}

	projectID := os.Getenv("PUBSUB_PROJECT_ID")
	topicID := "send-email"

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Erro ao criar client: %v", err)
	}

	topic = client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Erro ao checar tópico: %v", err)
	}
	if !exists {
		topic, _ = client.CreateTopic(ctx, topicID)
	}

	http.HandleFunc("/send-email", sendEmailHandler)

	port := os.Getenv("HOST")
	fmt.Println("API rodando na porta", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Apenas POST é permitido")
		return
	}

	var payload EmailPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Payload inválido:", err)
		return
	}

	data, _ := json.Marshal(payload)
	result := topic.Publish(context.Background(), &pubsub.Message{Data: data})
	id, err := result.Get(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erro ao publicar mensagem:", err)
		return
	}

	fmt.Fprintf(w, "Mensagem publicada com ID: %s\n", id)
}
