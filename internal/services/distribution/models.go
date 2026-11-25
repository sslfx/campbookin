package distribution

import "time"

type Participant struct {
	ID           string    `json:"id"`
	FullName     string    `json:"full_name"`
	Gender       string    `json:"gender"`
	BirthDate    time.Time `json:"birth_date"`
	Age          int       `json:"age"`
	ParentName   string    `json:"parent_name"`
	ParentPhone  string    `json:"parent_phone"`
	Organization string    `json:"organization"` // Организация (может быть "")
	ShiftID      string    `json:"shift_id"`
	SquadID      string    `json:"squad_id"`
}
type Squad struct {
	ID           string
	Participants []*Participant
	BoysCount    int
	GirlsCount   int
	AvgAge       float64
	MaxAge       float64
	MinAge       float64
	Capacity     int
}
type Room struct {
	ID           string         `json:"id"`
	Number       int            `json:"name"`
	MinAge       int            `json:"min_age"`
	MaxAge       int            `json:"max_age"`
	Capacity     int            `json:"capacity"`
	Occupied     int            `json:"occupied"`
	Participants []*Participant `json:"participants"`
}
type Building struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	FloorsCount int      `json:"floor"`
	Capacity    int      `json:"capacity"`
	Floors      []*Floor `json:"floors"`
}
type Floor struct {
	ID               string  `json:"id"`
	Number           int     `json:"number"`
	StartRoomsNumber int     `json:"start_rooms"`
	Rooms            []*Room `json:"rooms"`
	AllowedGender    string  `json:"gender"`
	CurrentGender    string  `json:"current_gender"`
	ReservedGender   string  `json:"reserved_gender"`
}
