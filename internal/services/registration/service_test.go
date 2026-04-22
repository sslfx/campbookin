package registration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	CreateUserFunc   func(ctx context.Context, user *User) error
	GetUserFunc      func(ctx context.Context, id string) (*User, error)
	GetUserEmailFunc func(ctx context.Context, email string) (*User, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetUser(ctx context.Context, id string) (*User, error) {
	return m.GetUserFunc(ctx, id)
}

func (m *mockUserRepo) GetUserEmail(ctx context.Context, email string) (*User, error) {
	return m.GetUserEmailFunc(ctx, email)
}

func TestRegister_Success(t *testing.T) {

	mockRepo := &mockUserRepo{
		CreateUserFunc: func(ctx context.Context, user *User) error {
			return nil
		},
	}
	secret := []byte("my-super-secret")
	service := NewService(mockRepo, secret)

	user, err := service.Register(context.Background(), "admin@zarnitsa.ru", "password123", "Max")

	if err != nil {
		t.Fatalf("Ожидался успех, но получили ошибку: %v", err)
	}
	if user == nil {
		t.Fatal("Ожидался объект пользователя, но получили nil")
	}
	if user.Email != "admin@zarnitsa.ru" {
		t.Errorf("Ожидался email admin@zarnitsa.ru, получили %s", user.Email)
	}
	if user.Role != "USER" {
		t.Errorf("Ожидалась роль USER, получили %s", user.Role)
	}
	if user.PasswordHash == "password123" {
		t.Error("Пароль не был захэширован!")
	}
}

func TestLogin_Success(t *testing.T) {

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	mockRepo := &mockUserRepo{
		GetUserEmailFunc: func(ctx context.Context, email string) (*User, error) {

			return &User{
				ID:           "user-123",
				Email:        email,
				PasswordHash: string(hash),
			}, nil
		},
	}

	secret := []byte("my-super-secret")
	service := NewService(mockRepo, secret)

	token, err := service.Login(context.Background(), "admin@zarnitsa.ru", "password123")

	if err != nil {
		t.Fatalf("Ожидался успех, но получили ошибку: %v", err)
	}
	if token == "" {
		t.Fatal("Ожидался токен, но получили пустую строку")
	}
}

func TestService_Errors(t *testing.T) {
	secret := []byte("secret")

	t.Run("RegEmptyData", func(t *testing.T) {
		mockRepo := &mockUserRepo{}
		service := NewService(mockRepo, []byte("secret"))

		user, err := service.Register(context.Background(), "", "pass", "Max")

		if err == nil {
			t.Error("Ожидалась ошибка из-за пустого email, но её нет")
		}
		if user != nil {
			t.Error("Пользователь не должен был создаться")
		}
	})

	t.Run("LoginEmptyData", func(t *testing.T) {
		mockRepo := &mockUserRepo{}
		service := NewService(mockRepo, []byte("my-super-secret"))

		token, err := service.Login(context.Background(), "", "")

		if err == nil {
			t.Fatal("Ожидалась ошибка из-за пустых данных, но её нет")
		}
		if token != "" {
			t.Fatal("Токен не должен был сгенерироваться")
		}
	})
	t.Run("RegisterBcryptError", func(t *testing.T) {
		longPass := string(make([]byte, 100))
		s := NewService(&mockUserRepo{}, secret)
		_, err := s.Register(context.Background(), "a@b.c", longPass, "Max")
		if err == nil {
			t.Error("Ожидалась ошибка RegErr")
		}
	})

	t.Run("RegisterRepoError", func(t *testing.T) {
		mockRepo := &mockUserRepo{
			CreateUserFunc: func(ctx context.Context, user *User) error {
				return errors.New("db crash")
			},
		}
		s := NewService(mockRepo, secret)
		_, err := s.Register(context.Background(), "a@b.c", "123", "Max")
		if err == nil || !strings.Contains(err.Error(), "create:") {
			t.Error("Ожидалась ошибка репозитория")
		}
	})

	t.Run("LoginUserNotFound", func(t *testing.T) {
		mockRepo := &mockUserRepo{
			GetUserEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, pgx.ErrNoRows
			},
		}
		s := NewService(mockRepo, secret)
		_, err := s.Login(context.Background(), "test@test.com", "123")
		if err == nil || err.Error() != "invalid credentials" {
			t.Error("Должно быть invalid credentials")
		}
	})

	t.Run("LoginWrongPassword", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("right_pass"), 4)
		mockRepo := &mockUserRepo{
			GetUserEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return &User{PasswordHash: string(hash)}, nil
			},
		}
		s := NewService(mockRepo, secret)
		_, err := s.Login(context.Background(), "a@b.c", "wrong_pass")
		if err == nil || err.Error() != "invalid credentials" {
			t.Error("Ожидалась ошибка пароля")
		}
	})

	t.Run("GetUserAllPaths", func(t *testing.T) {
		mockRepo := &mockUserRepo{
			GetUserFunc: func(ctx context.Context, id string) (*User, error) {
				if id == "found" {
					return &User{ID: "found"}, nil
				}
				if id == "not_found" {
					return nil, pgx.ErrNoRows
				}
				return nil, errors.New("db error")
			},
		}
		s := NewService(mockRepo, secret)

		_, err := s.GetUser(context.Background(), "")
		if err == nil {
			t.Error("Ожидалась ошибка на пустой ID")
		}

		_, err = s.GetUser(context.Background(), "found")
		if err != nil {
			t.Error("Должен найти юзера")
		}
		
		_, err = s.GetUser(context.Background(), "not_found")
		if err == nil {
			t.Error("Ожидалась ошибка 'invalid id'")
		}

		_, err = s.GetUser(context.Background(), "error")
		if err == nil {
			t.Error("Ожидалась ошибка GetUser:")
		}
	})
}
