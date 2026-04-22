package admin

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userIDKey contextKey = "userID"

func Interceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		publicMethods := []string{
			"GetShift",
			"GetCamp",
		}
		for _, method := range publicMethods {
			if strings.HasSuffix(info.FullMethod, method) {
				return handler(ctx, req)
			}
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}
		userIDs := md.Get("user_id")
		if len(userIDs) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing credentials")
		}
		ctx = context.WithValue(ctx, userIDKey, userIDs[0])
		return handler(ctx, req)
	}
}

func GetUserIDFromContext(ctx context.Context) string {
	value := ctx.Value("userID")
	if value == nil {
		return ""
	}
	return value.(string)
}
