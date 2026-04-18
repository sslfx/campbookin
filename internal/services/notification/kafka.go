package notification

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

func (n *NotificationService) StartKafkaListener(ctx context.Context, email, pass string) {
	// Используем новый GroupID, чтобы подхватить сообщения, которые могли быть "пропущены"
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "bookings.create",
		//GroupID:     "notification-service-v4-а",
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	log.Println(">>> Notification service: Listener started. Waiting for messages from Kafka...")

	job := make(chan BookingTicket, 100)

	// Запускаем воркеры
	for w := 1; w <= 5; w++ {
		go n.worker(ctx, email, pass, job, w)
	}

	for {
		// Читаем сообщение
		log.Println("DEBUG: Reader is polling...")
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("!!! KAFKA READ ERROR: %v", err)
			if ctx.Err() != nil {
				log.Println("!!! Kafka listener stopped by context")
				close(job)
				return
			}
			log.Println("!!! Read message error:", err)
			continue
		}

		log.Printf("--> Kafka Message Received! Partition: %d, Offset: %d", msg.Partition, msg.Offset)

		// Декодируем сообщение (создаем новую переменную для каждой итерации)
		var event BookingTicket
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Println("!!! Unmarshal message error:", err)
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
