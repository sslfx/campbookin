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
	IsOwner(ctx context.Context, userID, campID string) (bool, error)
	Assing(ctx context.Context, userID, campID string) error
}
type Service struct {
	repo CampAdminRepository
}

func NewService(repo CampAdminRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Assign(ctx context.Context, userID, campID string) error {
	role := GetRoleFromContext(ctx)
	if userID == "" || campID == "" {
		return fmt.Errorf("userID or campID is empty")
	}
	if role != "SUPER_ADMIN" {
		return fmt.Errorf("AssignDirector: forbidden")
	}
	_, err := s.repo.GetCamp(ctx, campID)
	if err != nil {
		return fmt.Errorf("AssignDirector:GetCamp: %w", err)
	}
	err = s.repo.Assing(ctx, userID, campID)
	if err != nil {
		return fmt.Errorf("AssignDirector:repo: %w", err)
	}
	return nil

}

func (s *Service) authorize(ctx context.Context, action string, campID string) error {
	role := GetRoleFromContext(ctx)
	userID := GetUserIDFromContext(ctx)

	switch role {

	case "SUPER_ADMIN":
		// супер админ может всё
		return nil

	case "DIRECTOR":
		switch action {

		case ActionGetCamp, ActionUpdateCamp, ActionDeleteCamp:
			ok, err := s.repo.IsOwner(ctx, userID, campID)
			if err != nil {
				return fmt.Errorf("authorize:IsOwner: %w", err)
			}
			if !ok {
				return fmt.Errorf("authorize: forbidden")
			}
			return nil

		case ActionListCamps:
			// список разрешён — но только свои лагеря (логика уже в repo)
			return nil

		default:
			return fmt.Errorf("authorize: forbidden for director")
		}

	default:
		return fmt.Errorf("authorize: unknown role")
	}
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
	role := GetRoleFromContext(ctx)
	userID := GetUserIDFromContext(ctx)

	if err := s.authorize(ctx, ActionListCamps, ""); err != nil {
		return nil, err
	}

	if role == "SUPER_ADMIN" {
		return s.repo.ListCamps(ctx)
	}

	if role == "DIRECTOR" {
		return s.repo.ListCampsDirector(ctx, userID)
	}

	return nil, fmt.Errorf("forbidden")
}
