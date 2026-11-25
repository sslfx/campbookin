package distribution

import (
	"errors"
	"fmt"
	"io"
	"log"
)

type DistribRepository interface {
	RepoCreateRoom(room *Room) error
	RepoCreateBuilding(building *Building) error
	RepoCreateFloor(floor *Floor) error
}

type Service struct {
	repo DistribRepository
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateSquadsAndDistribute(file io.Reader, squadsCount int, shiftID string, b []*Building) error {

	p, err := s.ImportParticipants(file, shiftID)
	if err != nil {
		log.Println("error importing participants")
		return fmt.Errorf("error importing participants: %w", err)
	}
	sq, err := s.CreatingSquads(p, squadsCount)
	if err != nil {
		log.Println("error creating squads")
		return fmt.Errorf("error creating squads: %w", err)
	}
	err = s.Distribute(sq, b)
	if err != nil {
		log.Println("error distributing squads")
		return fmt.Errorf("error creating distribute: %w", err)
	}
	return nil

}
func (s *Service) CreateRoom(number, capacity, occupied, minAge, maxAge int) (*Room, error) {
	if number <= 0 || capacity <= 0 || minAge < 6 || maxAge <= 0 || maxAge >= 18 {
		log.Println("CreateRoom: invalid parameters")
		return nil, errors.New("invalid params")
	}
	room := &Room{
		Number:       number,
		Capacity:     capacity,
		Occupied:     occupied,
		MinAge:       minAge,
		MaxAge:       maxAge,
		Participants: nil,
	}
	//if err := s.repo.RepoCreateRoom(room); err != nil {
	//log.Println(err)
	//return nil, fmt.Errorf("create room: %w", err)
	//}
	return room, nil
}

func (s *Service) AddRooms(startNum, countRooms, capacity, minAge, maxAge int) ([]*Room, error) {
	rooms := make([]*Room, 0, countRooms)

	for i := 0; i < countRooms; i++ {
		roomNumber := startNum + i
		r, err := s.CreateRoom(roomNumber, capacity, 0, minAge, maxAge)
		if err != nil {
			return nil, fmt.Errorf("ошибка на комнате %d: %w", roomNumber, err)
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}

func (s *Service) CreateFloor(number int, rooms []*Room, gender string) (*Floor, error) {
	if number <= 0 || len(rooms) == 0 || gender == "" {
		log.Println("CreateFloor: invalid parameters")
		return nil, errors.New("invalid params")
	}

	floor := &Floor{
		Number:           number,
		StartRoomsNumber: rooms[0].Number,
		Rooms:            rooms,
		AllowedGender:    gender,
	}
	//if err := s.repo.RepoCreateFloor(floor); err != nil {
	//log.Println(err)
	//return nil, fmt.Errorf("create floor: %w", err)
	//}

	return floor, nil

}

func (s *Service) CreateBuilding(name string, floorsCount int, floors []*Floor) (*Building, error) {
	if name == "" || floorsCount <= 0 || floors == nil || len(floors) == 0 {
		log.Println("CreateBuilding: invalid parameters")
		return nil, errors.New("invalid params")
	}
	capacity := 0
	for _, floor := range floors {
		for _, room := range floor.Rooms {
			capacity += room.Capacity
		}
	}
	building := &Building{
		Name:        name,
		FloorsCount: floorsCount,
		Capacity:    capacity,
		Floors:      floors,
	}
	//if err := s.repo.RepoCreateBuilding(building); err != nil {
	//log.Println(err)
	//return nil, fmt.Errorf("create building: %w", err)
	//}
	return building, nil
}
