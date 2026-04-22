package main

import (
	pb2 "campbookin/api/proto/proto_output/admin_proto"
	pb "campbookin/api/proto/proto_output/reg_proto"
	"campbookin/internal/services/notification"
	"campbookin/pkg"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const retryPolicy = `{
    "methodConfig": [{
        "name": [{"service": ""}], 
        "retryPolicy": {
            "MaxAttempts": 4,
            "InitialBackoff": "0.1s",
            "MaxBackoff": "1s",
            "BackoffMultiplier": 2,
            "RetryableStatusCodes": ["UNAVAILABLE"]
        }
    }]
}`

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	//env
	if err := pkg.LoadEnv(".env"); err != nil {
		os.Exit(2)
	}
	pass := pkg.MustGet("PASS")
	email := pkg.MustGet("EMAIL")
	redisAddr := pkg.MustGet("REDIS_ADDR")

	log.Println("Notification service started")

	connAuth, err := grpc.NewClient(
		"reg:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		log.Fatal(err)
	}

	connAdmin, err := grpc.NewClient(
		"admin:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		log.Fatal(err)
	}

	authClient := pb.NewCampRegistrationClient(connAuth)
	adminClient := pb2.NewCampAdminClient(connAdmin)

	defer connAuth.Close()
	defer connAdmin.Close()

	ns := notification.NewNotificationService(authClient, adminClient, redisAddr)

	done := make(chan struct{})
	go func() {
		log.Println("kafka started")
		ns.StartKafkaListener(ctx, email, pass)
		close(done)
	}()
	<-ctx.Done()

	log.Println("shutting down notification service")

	select {
	case <-done:
		log.Println("All workers finished. Shutdown complete.")
	case <-time.After(15 * time.Second):
		log.Println("Shutdown timeout reached. Forcing exit.")
	}

	if err := ns.StopRedis(); err != nil {
		log.Printf("Redis close error: %v", err)
	}
}
