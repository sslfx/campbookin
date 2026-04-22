package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/segmentio/kafka-go"
)

const (
	//camps
	ActionCreateCamp     = "camp.create"          //admin
	ActionDeleteCamp     = "camp.delete"          //admin
	ActionUpdateCamp     = "camp.update"          //dir
	ActionGetCamp        = "camp.get"             //all
	ActionListCamps      = "camp.list"            //all
	ActionAssignDirector = "camp.assign_director" //admin
	//shifts
	ActionCreateShift    = "shift.create"    //dir
	ActionUpdateShift    = "shift.update"    //dir
	ActionDeleteShift    = "shift.delete"    //dir
	ActionGetShift       = "shift.get"       //all
	ActionListShifts     = "shift.list"      //admin
	ActionListCampShifts = "shift.list_camp" //all
	//booking
	ActionBookShift = "shift.book" //all
)

func (s *Service) authorize(ctx context.Context, action string, campID string) error {
	userID := GetUserIDFromContext(ctx)
	globalRole, err := s.repo.GetUserRole(ctx, userID)
	if err != nil {
		return err
	}
	if globalRole == "SUPER_ADMIN" {
		return nil
	}

	isDirector, _ := s.repo.IsDirector(ctx, userID, campID)
	if isDirector {
		if action == ActionUpdateCamp || action == ActionGetCamp || action == ActionListCamps || action == ActionCreateShift || action == ActionUpdateShift || action == ActionDeleteShift || action == ActionGetShift || action == ActionListCampShifts || action == ActionBookShift {
			return nil
		}
		return fmt.Errorf("authorize: forbidden for director")
	}
	if action == ActionGetCamp || action == ActionListCamps || action == ActionGetShift || action == ActionListCampShifts || action == ActionBookShift {
		return nil
	}
	return fmt.Errorf("forbidden")
}

type CampAdminRepository interface {
	CreateCamp(ctx context.Context, camp *Camp) error
	GetCamp(ctx context.Context, id string) (*Camp, error)
	DeleteCamp(ctx context.Context, id string) error
	ListCamps(ctx context.Context) ([]*Camp, error)
	ListCampsDirector(ctx context.Context, userID string) ([]*Camp, error)
	UpdateCamp(ctx context.Context, camp *Camp) error
	IsDirector(ctx context.Context, userID, campID string) (bool, error)
	Assing(ctx context.Context, userID, campID string) error
	GetUserRole(ctx context.Context, userID string) (string, error)
	CreateShift(ctx context.Context, shift *Shift) error
	GetShift(ctx context.Context, id string) (*Shift, error)
	CreateBooking(ctx context.Context, shiftID, userID string) error
	DeleteShift(ctx context.Context, id string) error
	UpdateShift(ctx context.Context, shift *Shift) error
	ListShifts(ctx context.Context) ([]*Shift, error)
	ListShiftsForCamp(ctx context.Context, campID string) ([]*Shift, error)
}
type Service struct {
	repo   CampAdminRepository
	writer *kafka.Writer
}

func NewService(repo CampAdminRepository) *Service {
	return &Service{
		repo: repo,
		writer: &kafka.Writer{
			Addr:     kafka.TCP("kafka:9092"),
			Topic:    "bookings.create",
			Balancer: &kafka.LeastBytes{},
		},
	}
}
func (s *Service) Shutdown() error {
	log.Println("Closing Kafka writer")
	return s.writer.Close()
}
func (s *Service) CreateCamp(ctx context.Context, name, description, address string) (*Camp, error) {
	if err := s.authorize(ctx, ActionCreateCamp, ""); err != nil {
		return nil, err
	}

	if name == "" || description == "" || address == "" {
		return nil, fmt.Errorf("invalid input")
	}

	camp := &Camp{
		Name:        name,
		Description: description,
		Address:     address,
	}

	if err := s.repo.CreateCamp(ctx, camp); err != nil {
		return nil, fmt.Errorf("CreateCampServ: %w", err)
	}

	return camp, nil
}

func (s *Service) GetCamp(ctx context.Context, id string) (*Camp, error) {
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}
	return s.repo.GetCamp(ctx, id)
}

func (s *Service) DeleteCamp(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is empty")
	}

	if err := s.authorize(ctx, ActionDeleteCamp, id); err != nil {
		return err
	}

	return s.repo.DeleteCamp(ctx, id)
}

func (s *Service) UpdateCamp(ctx context.Context, id, name, description, address string) error {
	if id == "" || name == "" || description == "" || address == "" {
		return fmt.Errorf("invalid input")
	}

	if err := s.authorize(ctx, ActionUpdateCamp, id); err != nil {
		return err
	}

	camp := &Camp{
		ID:          id,
		Name:        name,
		Description: description,
		Address:     address,
	}

	return s.repo.UpdateCamp(ctx, camp)
}

func (s *Service) ListCamps(ctx context.Context) ([]*Camp, error) {
	return s.repo.ListCamps(ctx)
}

func (s *Service) Assign(ctx context.Context, userID, campID string) error {
	if userID == "" || campID == "" {
		return fmt.Errorf("invalid input")
	}
	if err := s.authorize(ctx, ActionAssignDirector, campID); err != nil {
		return err
	}
	if _, err := s.repo.GetCamp(ctx, campID); err != nil {
		return fmt.Errorf("camp not found: %w", err)
	}
	return s.repo.Assing(ctx, userID, campID)
}

func (s *Service) CreateShift(ctx context.Context, name, campID, description string, capacity, price int32, status ShiftStatus, startTime, endTime time.Time) (*Shift, error) {
	if err := s.authorize(ctx, ActionCreateShift, campID); err != nil {
		return nil, err
	}
	if name == "" || description == "" || campID == "" || capacity == 0 || price == 0 || startTime.IsZero() || endTime.IsZero() {
		return nil, fmt.Errorf("invalid input")
	}
	shift := &Shift{
		Name:        name,
		Description: description,
		Capacity:    capacity,
		Price:       price,
		Status:      status,
		StartTime:   startTime,
		EndTime:     endTime,
		CampID:      campID,
	}
	if err := s.repo.CreateShift(ctx, shift); err != nil {
		return nil, fmt.Errorf("CreateShiftServ: %w", err)
	}
	return shift, nil
}

func (s *Service) GetShift(ctx context.Context, id string) (*Shift, error) {
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}
	shift, err := s.repo.GetShift(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetShiftServ: %w", err)
	}
	return shift, nil
}
func (s *Service) UpdateShift(ctx context.Context, id, name, description string, startTime, endTime time.Time, capacity, price int32, status ShiftStatus) error {
	if id == "" || name == "" || description == "" || capacity == 0 || price == 0 || startTime.IsZero() || endTime.IsZero() {
		return fmt.Errorf("invalid input")
	}
	shift, err := s.repo.GetShift(ctx, id)
	if err != nil {
		return err
	}
	if err = s.authorize(ctx, ActionUpdateShift, shift.CampID); err != nil {
		return err
	}
	if err := s.repo.UpdateShift(ctx, &Shift{
		ID:          id,
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Capacity:    capacity,
		Price:       price,
		Status:      status,
	}); err != nil {
		return fmt.Errorf("UpdateShiftServ: %w", err)
	}

	return nil
}
func (s *Service) DeleteShift(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is empty")
	}
	shift, err := s.repo.GetShift(ctx, id)
	if err != nil {
		return err
	}
	err = s.authorize(ctx, ActionDeleteShift, shift.CampID)
	if err != nil {
		return err
	}
	err = s.repo.DeleteShift(ctx, id)
	if err != nil {
		return fmt.Errorf("DeleteShiftServ: %w", err)
	}
	return nil
}

func (s *Service) ListShifts(ctx context.Context) ([]*Shift, error) {
	err := s.authorize(ctx, ActionListShifts, "")
	if err != nil {
		return nil, err
	}
	shifts, err := s.repo.ListShifts(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListShiftsServ: %w", err)
	}
	return shifts, nil
}

func (s *Service) ListShiftsForCamp(ctx context.Context, campID string) ([]*Shift, error) {
	shifts, err := s.repo.ListShiftsForCamp(ctx, campID)
	if err != nil {
		return nil, fmt.Errorf("ListShiftsServ: %w", err)
	}
	return shifts, nil
}

func (s *Service) BookShift(ctx context.Context, shiftID, userID string) (bool, error) {
	if userID == "" || shiftID == "" {
		return false, fmt.Errorf("invalid input")
	}
	shift, err := s.repo.GetShift(ctx, shiftID)
	if err != nil {
		return false, fmt.Errorf("GetShiftServ: %w", err)
	}
	if shift.EndTime.Before(time.Now()) {
		return false, fmt.Errorf("shift is already active or expired")
	}
	if err := s.repo.CreateBooking(ctx, shift.ID, userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return false, fmt.Errorf("already booking")
			}
		}
		return false, fmt.Errorf("CreateBookingServ: %w", err)
	}

	type BookingEvent struct {
		UserID  string `json:"user_id"`
		ShiftID string `json:"shift_id"`
	}
	event := BookingEvent{
		UserID:  userID,
		ShiftID: shiftID,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return false, fmt.Errorf("CreateBooking: %w", err)
	}
	if err := s.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	}); err != nil {
		fmt.Printf("ERROR: Kafka write failed: %v\n", err)
		return false, fmt.Errorf("CreateBookingKafka: %w", err)
	}

	fmt.Println("DEBUG: Message successfully sent to Kafka!")
	return true, nil
}
