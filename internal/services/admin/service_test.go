package admin

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type mockAdminRepo struct {
	CampAdminRepository
	GetUserRoleFunc   func(ctx context.Context, userID string) (string, error)
	IsDirectorFunc    func(ctx context.Context, userID, campID string) (bool, error)
	CreateCampFunc    func(ctx context.Context, camp *Camp) error
	CreateShiftFunc   func(ctx context.Context, shift *Shift) error
	GetShiftFunc      func(ctx context.Context, id string) (*Shift, error)
	CreateBookingFunc func(ctx context.Context, shiftID, userID string) error
	DeleteShiftFunc   func(ctx context.Context, id string) error
}

func (m *mockAdminRepo) GetUserRole(ctx context.Context, id string) (string, error) {
	return m.GetUserRoleFunc(ctx, id)
}
func (m *mockAdminRepo) IsDirector(ctx context.Context, u, c string) (bool, error) {
	return m.IsDirectorFunc(ctx, u, c)
}
func (m *mockAdminRepo) CreateCamp(ctx context.Context, c *Camp) error {
	return m.CreateCampFunc(ctx, c)
}
func (m *mockAdminRepo) CreateShift(ctx context.Context, s *Shift) error {
	return m.CreateShiftFunc(ctx, s)
}
func (m *mockAdminRepo) GetShift(ctx context.Context, id string) (*Shift, error) {
	return m.GetShiftFunc(ctx, id)
}
func (m *mockAdminRepo) CreateBooking(ctx context.Context, s, u string) error {
	return m.CreateBookingFunc(ctx, s, u)
}
func (m *mockAdminRepo) DeleteShift(ctx context.Context, id string) error {
	return m.DeleteShiftFunc(ctx, id)
}

func TestCreateCampAsAdmin(t *testing.T) {
	mockRepo := &mockAdminRepo{
		GetUserRoleFunc: func(ctx context.Context, id string) (string, error) {
			return "SUPER_ADMIN", nil
		},
		CreateCampFunc: func(ctx context.Context, camp *Camp) error {
			return nil
		},
	}
	s := &Service{repo: mockRepo}

	ctx := context.WithValue(context.Background(), userIDKey, "admin-1")

	camp, err := s.CreateCamp(ctx, "Зарница", "Описание", "Ржев")

	if err != nil {
		t.Fatalf("Админ должен иметь возможность создать лагерь, но получили ошибку: %v", err)
	}
	if camp.Name != "Зарница" {
		t.Errorf("Ожидалось имя Зарница, получили %s", camp.Name)
	}
}

func TestCreateShiftAsDirector(t *testing.T) {
	mockRepo := &mockAdminRepo{
		GetUserRoleFunc: func(ctx context.Context, id string) (string, error) {
			return "USER", nil // Глобальная роль просто юзер
		},
		IsDirectorFunc: func(ctx context.Context, u, c string) (bool, error) {
			return true, nil // Но он директор этого лагеря
		},
		CreateShiftFunc: func(ctx context.Context, shift *Shift) error {
			return nil
		},
	}
	s := &Service{repo: mockRepo}
	ctx := context.WithValue(context.Background(), userIDKey, "dir-1")

	startTime := time.Now().Add(time.Hour)
	endTime := time.Now().Add(time.Hour * 24)

	_, err := s.CreateShift(ctx, "Первая смена", "camp-123", "Инфо", 100, 5000, 1, startTime, endTime)

	if err != nil {
		t.Errorf("Директор лагеря должен иметь возможность создать смену: %v", err)
	}
}
func TestDeleteShiftSuccess(t *testing.T) {
	mockRepo := &mockAdminRepo{
		GetShiftFunc: func(ctx context.Context, id string) (*Shift, error) {
			return &Shift{ID: id, CampID: "c1"}, nil
		},
		GetUserRoleFunc: func(ctx context.Context, id string) (string, error) {
			return "SUPER_ADMIN", nil
		},
		DeleteShiftFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}
	s := &Service{repo: mockRepo}
	ctx := context.WithValue(context.Background(), userIDKey, "admin")

	err := s.DeleteShift(ctx, "shift-123")
	if err != nil {
		t.Errorf("Ошибка при удалении смены: %v", err)
	}
}
func TestBookShiftValidation(t *testing.T) {
	mockRepo := &mockAdminRepo{
		GetShiftFunc: func(ctx context.Context, id string) (*Shift, error) {
			return &Shift{
				ID:      "shift-1",
				EndTime: time.Now().Add(-time.Hour),
			}, nil
		},
	}
	s := &Service{repo: mockRepo}

	success, err := s.BookShift(context.Background(), "shift-1", "user-1")

	if success || err == nil {
		t.Error("Ожидалась ошибка: нельзя бронировать прошедшую смену")
	}
}

func TestAuthorizeForbidden(t *testing.T) {
	mockRepo := &mockAdminRepo{
		GetUserRoleFunc: func(ctx context.Context, id string) (string, error) {
			return "USER", nil
		},
		IsDirectorFunc: func(ctx context.Context, u, c string) (bool, error) {
			return false, nil
		},
	}
	s := &Service{repo: mockRepo}
	ctx := context.WithValue(context.Background(), userIDKey, "simple-user")

	err := s.authorize(ctx, ActionDeleteCamp, "camp-1")

	if err == nil {
		t.Error("Ожидалась ошибка доступа (Forbidden), но доступ разрешен")
	}
}

func TestAuthorizeDirectorLimits(t *testing.T) {
	mockRepo := &mockAdminRepo{
		GetUserRoleFunc: func(ctx context.Context, id string) (string, error) {
			return "USER", nil
		},
		IsDirectorFunc: func(ctx context.Context, u, c string) (bool, error) {
			return true, nil
		},
	}
	s := &Service{repo: mockRepo}
	ctx := context.WithValue(context.Background(), userIDKey, "dir-1")

	err := s.authorize(ctx, ActionDeleteCamp, "camp-1")
	if err == nil {
		t.Error("Ожидалась ошибка: директор не должен иметь прав на удаление лагеря")
	}

	err = s.authorize(ctx, ActionUpdateShift, "camp-1")
	if err != nil {
		t.Errorf("Директор должен иметь право на обновление смены: %v", err)
	}
}
func TestBookShiftKafkaError(t *testing.T) {
	mockRepo := &mockAdminRepo{
		GetShiftFunc: func(ctx context.Context, id string) (*Shift, error) {
			return &Shift{ID: "s1", EndTime: time.Now().Add(time.Hour)}, nil
		},
		CreateBookingFunc: func(ctx context.Context, s, u string) error {
			return nil
		},
	}
	s := NewService(mockRepo)
	_, err := s.BookShift(context.Background(), "s1", "u1")
	if err == nil {
		t.Log("Если Kafka не запущена, здесь должна быть ошибка записи")
	}
}
func TestBookShiftAlreadyBooked(t *testing.T) {
	mockRepo := &mockAdminRepo{
		GetShiftFunc: func(ctx context.Context, id string) (*Shift, error) {
			return &Shift{ID: "s1", EndTime: time.Now().Add(time.Hour)}, nil
		},
		CreateBookingFunc: func(ctx context.Context, s, u string) error {
			return &pgconn.PgError{Code: "23505"}
		},
	}
	s := NewService(mockRepo)
	_, err := s.BookShift(context.Background(), "s1", "u1")
	if err == nil || err.Error() != "already booking" {
		t.Errorf("Ожидалась ошибка 'already booking', получили: %v", err)
	}
}
