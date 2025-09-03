package main

import (
	"context"
	"log"
	"os"
	"path"

	"order_service/config"

	"github.com/segmentio/kafka-go"
)

const orderDir = "orders"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
	})
	defer writer.Close() //nolint:errcheck

	workDir, err := os.Getwd()
	if err != nil {
		log.Println("failed to read dir:", err)
		return
	}
	dirEntry, err := os.ReadDir(path.Join(workDir, orderDir))
	if err != nil {
		log.Println("failed to read dir:", err)
		return
	}

	msgs := make([]kafka.Message, 0, len(dirEntry))
	for _, entry := range dirEntry {
		data, err := os.ReadFile(path.Join(workDir, orderDir, entry.Name())) // #nosec G304
		if err != nil {
			log.Println("failed to read file:", err)
			return
		}
		msgs = append(msgs, kafka.Message{Value: data})
	}

	err = writer.WriteMessages(context.Background(), msgs...)
	if err != nil {
		log.Println("failed to send message:", err)
		return
	}

	log.Printf("Message sent to topic: %s successfully!\n", cfg.Topic)
}
