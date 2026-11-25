package notification

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"
)

func (n *NotificationService) StartKafkaListener(ctx context.Context, email, pass string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"kafka:9092"},
		Topic:       "bookings.create",
		GroupID:     "notification-service",
		StartOffset: kafka.FirstOffset,
	})
	defer func() { _ = reader.Close() }()

	log.Println("Waiting for messages from Kafka")

	job := make(chan BookingTicket, 100)
	var wg sync.WaitGroup
	// Запускаем воркеры
	for w := 1; w <= 5; w++ {
		wg.Add(1)
		go func() {
			n.worker(ctx, email, pass, job, w)
			wg.Done()
		}()
	}

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Println("Error reading from Kafka:", err)
			break
		}
		var event BookingTicket
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}
		job <- event
	}

	close(job)
	wg.Wait()
	log.Println("Shutting down Kafka listener")
}

func (n *NotificationService) worker(ctx context.Context, email, pass string, jobs <-chan BookingTicket, workerID int) {
	for event := range jobs {
		log.Printf("[Worker %d] Starting PDF generation for User", workerID)
		n.processBooking(ctx, email, pass, &event)
		log.Printf("[Worker %d] PDF process finished", workerID)
	}
}
