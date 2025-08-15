package main

import (
	"context"
	"log"
	"order_service/config"
	"os"
	"path"

	"github.com/segmentio/kafka-go"
)

const orderDir = "orders"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	})
	defer writer.Close()

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalln("failed to read dir:", err)
	}
	dirEntry, err := os.ReadDir(path.Join(workDir, orderDir))
	if err != nil {
		log.Fatalln("failed to read dir:", err)
	}

	msgs := make([]kafka.Message, 0, len(dirEntry))
	for _, entry := range dirEntry {
		data, err := os.ReadFile(path.Join(workDir, orderDir, entry.Name()))
		if err != nil {
			log.Fatalln("failed to read file:", err)
		}
		msgs = append(msgs, kafka.Message{Value: data})
	}

	err = writer.WriteMessages(context.Background(), msgs...)
	if err != nil {
		log.Fatalln("failed to send message:", err)
	}

	log.Printf("Message sent to topic: %s successfully!\n", cfg.Kafka.Topic)
}
