package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/bissquit/url-shortener/internal/auth"
	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/service"
	pb "github.com/bissquit/url-shortener/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ShortenerServer implements pb.ShortenerServiceServer.
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	storage   repository.URLRepository
	baseURL   string
	generator service.IDGenerator
}

// NewShortenerServer creates a new gRPC shortener server.
func NewShortenerServer(storage repository.URLRepository, baseURL string, generator service.IDGenerator) *ShortenerServer {
	return &ShortenerServer{
		storage:   storage,
		baseURL:   baseURL,
		generator: generator,
	}
}

// ShortenURL handles URL shortening (analog of POST /api/shorten).
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "empty URL")
	}

	userID, _ := auth.GetUserIDFromContext(ctx)

	shortURL, err := s.generateAndStoreShortURL(req.Url, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.URLShortenResponse{Result: shortURL}, nil
}

// ExpandURL handles URL expansion (analog of GET /{id}).
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "empty ID")
	}

	originalURL, err := s.storage.GetURLByID(req.Id)
	if errors.Is(err, repository.ErrDeleted) {
		return nil, status.Error(codes.NotFound, "URL was deleted")
	}
	if err != nil {
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	return &pb.URLExpandResponse{Result: originalURL}, nil
}

// ListUserURLs returns all user's URLs (analog of GET /api/user/urls).
func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, ok := auth.GetUserIDFromContext(ctx)
	if userID == "" || !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	items, err := s.storage.GetURLsByUserID(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.UserURLsResponse{}
	for _, it := range items {
		shortURL, err := url.JoinPath(s.baseURL, it.ShortID)
		if err != nil {
			log.Printf("ERROR: cannot build short url for %s: %v", it.ShortID, err)
			continue
		}
		resp.Url = append(resp.Url, &pb.URLData{
			ShortUrl:    shortURL,
			OriginalUrl: it.OriginalURL,
		})
	}

	return resp, nil
}

func (s *ShortenerServer) generateAndStoreShortURL(originalURL, userID string) (string, error) {
	const maxAttempts = 10

	for i := 0; i < maxAttempts; i++ {
		id, err := s.generator.GenerateShortID()
		if err != nil {
			return "", fmt.Errorf("cannot generate shorten ID: %w", err)
		}

		err = s.storage.Create(id, originalURL, userID)
		switch {
		case err == nil:
			return url.JoinPath(s.baseURL, id)
		case errors.Is(err, repository.ErrIDAlreadyExists):
			continue
		case errors.Is(err, repository.ErrURLAlreadyExists):
			existingID, err2 := s.storage.GetIDByURL(originalURL)
			if err2 != nil {
				return "", err2
			}
			return url.JoinPath(s.baseURL, existingID)
		default:
			return "", err
		}
	}

	return "", fmt.Errorf("id generation exhausted")
}
