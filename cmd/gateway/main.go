package main

import (
	pb2 "campbookin/api/proto/proto_output/admin_proto"
	pb1 "campbookin/api/proto/proto_output/reg_proto"
	"campbookin/internal/services/api-gateway"
	"campbookin/pkg"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
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

	if err := pkg.LoadEnv(".env"); err != nil {
		os.Exit(1)
	}
	secret := pkg.MustGet("PASS_SECRET")
	addr := pkg.MustGet("REDIS_ADDR")
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	connAuth, err := grpc.NewClient("reg:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer connAuth.Close()
	connAdmin, err := grpc.NewClient("admin:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer connAdmin.Close()
	authClient := pb1.NewCampRegistrationClient(connAuth)
	adminClient := pb2.NewCampAdminClient(connAdmin)
	gw := api_gateway.NewGateway(authClient, adminClient, []byte(secret), addr)
	r := chi.NewRouter()
	gw.RegisterRoutes(r)
	server := http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	log.Printf("Listening on port 8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen error: %v", err)
	}
}
