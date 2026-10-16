package grpcserver

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"mtv-erp/auth-service/internal/auth"
	authpb "mtv-erp/auth-service/internal/authpb/auth/v1"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Unauthenticated, "credenciais inválidas")
		}
		return nil, status.Errorf(codes.Internal, "erro ao buscar usuário: %v", err)
	}

	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, status.Error(codes.Unauthenticated, "credenciais inválidas")
	}

	token, err := auth.GenerateToken(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao gerar token: %v", err)
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