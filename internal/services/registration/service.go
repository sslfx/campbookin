package registration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserEmail(ctx context.Context, email string) (*User, error)
}
type Service struct {
	repo   UserRepo
	secret []byte
}

func NewService(repo UserRepo, secret []byte) *Service {
	return &Service{repo: repo, secret: secret}
}

func (s *Service) generateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*User, error) {
	if email == "" || password == "" || name == "" {
		return nil, fmt.Errorf("email or password is empty")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("RegErr: %w", err)
	}
	user := &User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         "USER",
		Name:         name,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	return user, nil

}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	if email == "" || password == "" {
		return "", fmt.Errorf("email or password is empty")
	}
	user, err := s.repo.GetUserEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return "", fmt.Errorf("Login: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}
	tokenString, err := s.generateToken(user)
	if err != nil {
		return "", fmt.Errorf("LoginErr: %w", err)
	}
	return tokenString, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("invalid id")
	}
	user, err := s.repo.GetUser(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("invalid id")
	}
	if err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}
	return user, nil
}
