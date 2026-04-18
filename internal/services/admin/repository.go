package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func generageID() string {
	id := uuid.New().String()
	return id
}
func (r *Repository) CreateCamp(ctx context.Context, camp *Camp) error {
	if camp.ID == "" {
		id := generageID()
		camp.ID = id
	}
	query := `INSERT INTO camps (id, name, description, address) VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, camp.ID, camp.Name, camp.Description, camp.Address)
	if err != nil {
		return fmt.Errorf("CreateCameRep: %w", err)
	}
	return nil
}

func (r *Repository) DeleteCamp(ctx context.Context, id string) error {
	query := `DELETE FROM camps WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DeleteCamp: %w", err)
	}
	return nil
}

func (r *Repository) GetCamp(ctx context.Context, id string) (*Camp, error) {
	if id == "" {
		return nil, fmt.Errorf("GetCampRep: empty id")
	}
	query := `SELECT id,name,description,address FROM camps WHERE id = $1`
	var camp Camp
	err := r.pool.QueryRow(ctx, query, id).Scan(&camp.ID, &camp.Name, &camp.Description, &camp.Address)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("GetCampRep:not found")
	}
	if err != nil {
		return nil, fmt.Errorf("GetCampRep: %w", err)
	}
	return &camp, nil
}

func (r *Repository) ListCamps(ctx context.Context) ([]*Camp, error) {
	query := `SELECT id,name,description,address FROM camps`
	var camps []*Camp
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ListCamps: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var camp Camp
		if err := rows.Scan(&camp.ID, &camp.Name, &camp.Description, &camp.Address); err != nil {
			return nil, fmt.Errorf("ListCamps: %w", err)
		}
		camps = append(camps, &camp)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListCamps: %w", err)
	}
	return camps, nil
}

func (r *Repository) ListCampsDirector(ctx context.Context, userID string) ([]*Camp, error) {
	if userID == "" {
		return nil, fmt.Errorf("unauthenticated")
	}
	query := `
        SELECT c.id, c.name, c.description, c.address
        FROM camps c
        JOIN camp_users cu ON cu.camp_id = c.id
        WHERE cu.user_id = $1
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ListCamps: %w", err)
	}
	defer rows.Close()
	var camps []*Camp
	for rows.Next() {
		var camp Camp
		if err := rows.Scan(&camp.ID, &camp.Name, &camp.Description, &camp.Address); err != nil {
			return nil, err
		}
		camps = append(camps, &camp)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListCamps: %w", err)
	}
	return camps, nil
}

func (r *Repository) UpdateCamp(ctx context.Context, camp *Camp) error {
	query := `UPDATE camps SET name = $1, description = $2, address = $3 WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, camp.Name, camp.Description, camp.Address, camp.ID)
	if err != nil {
		return fmt.Errorf("UpdateCampRep: %w", err)
	}
	return nil
}

func (r *Repository) Assing(ctx context.Context, userID, campID string) error {
	query := `
	INSERT INTO camp_users (user_id, camp_id, role)
	VALUES ($1, $2, 'DIRECTOR')
	ON CONFLICT (user_id, camp_id)
	DO NOTHING;
	`
	_, err := r.pool.Exec(ctx, query, userID, campID)
	if err != nil {
		return fmt.Errorf("Assing: %w", err)
	}
	return nil

}
func (r *Repository) GetUserRole(ctx context.Context, userID string) (string, error) {
	var role string
	err := r.pool.QueryRow(ctx,
		"SELECT role FROM users WHERE id = $1",
		userID,
	).Scan(&role)

	if err != nil {
		return "", err
	}

	return role, nil
}
func (r *Repository) IsDirector(ctx context.Context, userID, campID string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1 
            FROM camp_users 
            WHERE user_id = $1 
              AND camp_id = $2 
              AND role = 'DIRECTOR'
        )
    `

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, campID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("IsOwnerERR: %w", err)
	}

	return exists, nil
}

func (r *Repository) CreateShift(ctx context.Context, shift *Shift) error {
	if shift.ID == "" {
		id := generageID()
		shift.ID = id
	}
	query := `INSERT INTO shifts (id, camp_id, name, description,start_time, end_time, price, capacity, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query, shift.ID, shift.CampID, shift.Name, shift.Description, shift.StartTime, shift.EndTime, shift.Price, shift.Capacity, shift.Status)
	if err != nil {
		return fmt.Errorf("CreateShift: %w", err)
	}
	return nil
}

func (r *Repository) GetShift(ctx context.Context, id string) (*Shift, error) {
	if id == "" {
		return nil, fmt.Errorf("unauthenticated")
	}
	query := `SELECT id,camp_id,description,name,start_time,end_time,price,capacity,status FROM shifts WHERE id = $1`
	var shift Shift
	err := r.pool.QueryRow(ctx, query, id).Scan(&shift.ID, &shift.CampID, &shift.Description, &shift.Name, &shift.StartTime, &shift.EndTime, &shift.Price, &shift.Capacity, &shift.Status)
	if err != nil {
		return nil, fmt.Errorf("GetShift: %w", err)
	}
	return &shift, nil
}
func (r *Repository) CreateBooking(ctx context.Context, shiftID, userID string) error {
	id := generageID()
	query := `INSERT INTO bookings (id,user_id,shift_id) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, id, userID, shiftID)
	if err != nil {
		return fmt.Errorf("CreateBooking: %w", err)
	}
	return nil
}
