package registration

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func generateID() string {
	id := uuid.New().String()
	return id
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	if user.ID == "" {
		id := generateID()
		user.ID = id
	}
	query := `INSERT INTO users (id, email, password_hash, role, name) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role, user.Name)
	if err != nil {
		return fmt.Errorf("CreateUserRep: %w", err)
	}
	return nil
}

func (r *Repository) GetUser(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, password_hash, role, name FROM users WHERE id = $1`
	var user User
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}
	return &user, nil
}
func (r *Repository) GetUserEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password_hash, role, name FROM users WHERE email = $1`
	var user User
	err := r.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("GetUserRep:not found")
	}
	if err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}
	return &user, nil
}
