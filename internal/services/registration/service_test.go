package registration

import (
	"context"
	"testing"

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

func TestRegister_EmptyFields(t *testing.T) {

	mockRepo := &mockUserRepo{}
	service := NewService(mockRepo, []byte("secret"))

	user, err := service.Register(context.Background(), "", "pass", "Max")

	if err == nil {
		t.Error("Ожидалась ошибка из-за пустого email, но её нет")
	}
	if user != nil {
		t.Error("Пользователь не должен был создаться")
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

func TestLogin_Empty(t *testing.T) {
	mockRepo := &mockUserRepo{}
	service := NewService(mockRepo, []byte("my-super-secret"))

	token, err := service.Login(context.Background(), "", "")

	if err == nil {
		t.Fatal("Ожидалась ошибка из-за пустых данных, но её нет")
	}
	if token != "" {
		t.Fatal("Токен не должен был сгенерироваться")
	}
}
