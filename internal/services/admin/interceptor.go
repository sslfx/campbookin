package admin

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const (
	userIDKey ctxKey = "userID"
	roleKey   ctxKey = "role"
)

func Interceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}
		roles := md.Get("role")
		userIDs := md.Get("user_id")
		if len(roles) == 0 || len(userIDs) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing credentials")
		}
		ctx = context.WithValue(ctx, roleKey, roles[0])
		ctx = context.WithValue(ctx, userIDKey, userIDs[0])
		return handler(ctx, req)
	}
}

func GetRoleFromContext(ctx context.Context) string {
	value := ctx.Value(roleKey)
	if value == nil {
		return ""
	}
	return value.(string)
}
func GetUserIDFromContext(ctx context.Context) string {
	value := ctx.Value(userIDKey)
	if value == nil {
		return ""
	}
	return value.(string)
}
