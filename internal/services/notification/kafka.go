package notification

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

func (n *NotificationService) StartKafkaListener(ctx context.Context, email, pass string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "bookings.create",
		//GroupID:     "notification-service-v4-а",
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	log.Println("Waiting for messages from Kafka")

	job := make(chan BookingTicket, 100)

	// Запускаем воркеры
	for w := 1; w <= 5; w++ {
		go n.worker(ctx, email, pass, job, w)
	}

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				close(job)
				return
			}
			continue
		}
		var event BookingTicket
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}
		job <- event
	}
}

func (n *NotificationService) worker(ctx context.Context, email, pass string, jobs <-chan BookingTicket, workerID int) {
	for event := range jobs {
		log.Printf("[Worker %d] Starting PDF generation for User", workerID)
		n.processBooking(ctx, email, pass, &event)
		log.Printf("[Worker %d] PDF process finished", workerID)
	}
}
