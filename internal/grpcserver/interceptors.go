package grpcserver

import (
	"context"

	"github.com/bissquit/url-shortener/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var secretKey = []byte("my-secret-key-change-in-production")

// AuthInterceptor is a gRPC unary interceptor that handles JWT auth via metadata.
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var userID string

	// try to get token from metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("authorization"); len(vals) > 0 {
			token, err := jwt.ParseWithClaims(vals[0], &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*auth.Claims); ok {
					userID = claims.UserID
				}
			}
		}
	}

	// generate new user if no valid token
	if userID == "" {
		userID = uuid.New().String()
	}

	ctx = context.WithValue(ctx, auth.UserIDKey, userID)
	return handler(ctx, req)
}
