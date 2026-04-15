package admin

import (
	"context"
	"fmt"
)

const (
	ActionCreateCamp     = "camp.create"
	ActionUpdateCamp     = "camp.update"
	ActionDeleteCamp     = "camp.delete"
	ActionGetCamp        = "camp.get"
	ActionListCamps      = "camp.list"
	ActionAssignDirector = "camp.assign_director"
)

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
}
type Service struct {
	repo CampAdminRepository
}

func NewService(repo CampAdminRepository) *Service {
	return &Service{repo: repo}
}

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
		if action == ActionGetCamp || action == ActionUpdateCamp || action == ActionDeleteCamp {
			return nil
		}
		if action == ActionCreateCamp || action == ActionListCamps {
			return fmt.Errorf("authorize: forbidden for director")
		}
	}
	return fmt.Errorf("forbidden")
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

	if err := s.authorize(ctx, ActionGetCamp, id); err != nil {
		return nil, err
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
	err := s.authorize(ctx, ActionListCamps, "")
	if err != nil {
		return nil, fmt.Errorf("authorize:get role: %w", err)
	}
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
