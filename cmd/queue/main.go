package main

import (
    "context"
    "fmt"
    "log"
    "os"
	"encoding/json"

    "cloud.google.com/go/pubsub"
    "github.com/joho/godotenv"
)

func main() {
    // Carrega .env
    err := godotenv.Load()
    if err != nil {
        log.Println("Não foi possível carregar .env, usando variáveis do sistema")
    }

    ctx := context.Background()
    projectID := os.Getenv("PUBSUB_PROJECT_ID")
    topicID := "send-email"

    client, err := pubsub.NewClient(ctx, projectID)
    if err != nil {
        log.Fatalf("Erro ao criar client: %v", err)
    }

    // Cria tópico se não existir
    topic := client.Topic(topicID)
    exists, _ := topic.Exists(ctx)
    if !exists {
        topic, _ = client.CreateTopic(ctx, topicID)
    }

    // Cria subscription
    subID := "send-email-sub"
    sub := client.Subscription(subID)
    exists, _ = sub.Exists(ctx)
    if !exists {
        sub, _ = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
            Topic: topic,
        })
    }

    // Recebe mensagens
    sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
        fmt.Println("Mensagem recebida:", string(msg.Data))
		var data struct {
			To      string `json:"to"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		}
		if err := json.Unmarshal(msg.Data, &data); err == nil {
			fmt.Println("email enviado para:", data.To, "Mensagem:", data.Body, "Assunto:", data.Subject)
		}
        msg.Ack()
    })
}
