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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	//env
	if err := pkg.LoadEnv(".env"); err != nil {
		os.Exit(2)
	}

	pass := pkg.MustGet("PASS")
	email := pkg.MustGet("EMAIL")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Println("Notification service started")
	connAuth, err := grpc.NewClient(
		"reg:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer connAuth.Close()
	if err != nil {
		log.Fatal(err)
	}
	authClient := pb.NewCampRegistrationClient(connAuth)

	// admin
	connAdmin, err := grpc.NewClient(
		"admin:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer connAdmin.Close()
	if err != nil {
		log.Fatal(err)
	}
	adminClient := pb2.NewCampAdminClient(connAdmin)
	ns := notification.NewNotificationService(authClient, adminClient)
	//go func() {
	//ns.StartKafkaListener(ctx, email, pass)
	//}()
	//<-ctx.Done()
	log.Println(">>> Starting Kafka Listener...")
	ns.StartKafkaListener(ctx, email, pass)
	log.Println("Shutting down")
}
