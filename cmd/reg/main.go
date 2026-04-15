package main

import (
	pb "campbookin/api/proto/proto_output/reg_proto"
	"campbookin/internal/services/registration"
	"campbookin/pkg"
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	//env
	if err := pkg.LoadEnv(".env"); err != nil {
		os.Exit(1)
	}
	env := pkg.Get("LOCAL_ENV", "info")
	secret := pkg.MustGet("PASS_SECRET")
	dsn := pkg.MustGet("DB_URL")
	port := pkg.MustGet("PORT1")

	//logger
	pkg.Setup(env)
	logger := pkg.With()
	//db
	pool := pkg.MustNewPool(ctx, dsn)

	//deps
	repo := registration.NewRepository(pool)
	service := registration.NewService(repo, []byte(secret))
	handler := registration.NewHandler(service, logger)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCampRegistrationServer(grpcServer, handler)
	log.Printf("RegServer listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
