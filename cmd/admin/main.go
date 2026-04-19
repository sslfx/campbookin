package main

import (
	pb "campbookin/api/proto/proto_output/admin_proto"
	"campbookin/internal/services/admin"
	"campbookin/pkg"
	"context"
	"embed"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	ctx := context.Background()

	//env
	if err := pkg.LoadEnv(".env"); err != nil {
		os.Exit(2)
	}
	env := pkg.Get("LOCAL_ENV", "info")
	dsn := pkg.MustGet("DB_URL")
	port := pkg.MustGet("PORT2")

	//logger
	pkg.Setup(env)
	logger := pkg.With()
	//db
	pool := pkg.MustNewPool(ctx, dsn)
	sqlDB := stdlib.OpenDBFromPool(pool)
	log.Printf("Applying database migrations")
	mig, err := pkg.EmbedMigrations(sqlDB, embedMigrations, "migrations")
	if err != nil {
		logger.Error("migrations error", "error", err)
	}
	if err := mig.Up(); err != nil {
		logger.Error("migrations error", "error", err)
	}
	//deps
	repo := admin.NewRepository(pool)
	service := admin.NewService(repo)
	handler := admin.NewHandler(service, logger)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(admin.Interceptor()))
	pb.RegisterCampAdminServer(grpcServer, handler)
	log.Printf("RegServer listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
