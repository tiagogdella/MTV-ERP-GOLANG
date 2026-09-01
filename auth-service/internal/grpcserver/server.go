package grpcserver

import (
	"context"
	"errors"

	"mtv-erp/auth-service/internal/auth"
	"mtv-erp/auth-service/internal/authpb"
	"mtv-erp/auth-service/internal/db"
)

type Server struct {
	authpb.UnimplementedAuthServiceServer
	repo       *db.UserRepository
	jwtSecret  string
}

func NewServer(repo *db.UserRepository, jwtSecret string) *Server{
	return &Server{repo: repo, jwtSecret: jwtSecret}
}

func (s *Server) CreateUser(ctx context.Context, req *authpb.CreateUserRequest) (*authpb.CreateUserResponse, error) {
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &db.User{
		Email: 			req.Email,
		PasswordHash: 	hash,
		Role: 			req.Role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return &authpb.CreateUserResponse{UserId: user.ID}, nil
}

func (s *Server) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &authpb.LoginResponse{Token: token}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	claims, err := auth.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid: true,
		UserId: claims.Subject,
		Role: claims.Role,
	}, nil
}