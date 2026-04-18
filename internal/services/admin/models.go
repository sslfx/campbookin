package admin

import "time"

type Camp struct {
	ID          string
	Name        string
	Description string
	Address     string
}

type ShiftStatus int

const (
	ShiftStatusUnknown  ShiftStatus = iota // 0
	ShiftStatusActive                      // 1
	ShiftStatusInactive                    // 2
)

type Shift struct {
	ID          string
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Capacity    int32
	Price       int32
	CampID      string
	Status      ShiftStatus
}
