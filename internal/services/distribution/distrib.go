package distribution

import (
	"fmt"
	"log"
	"sort"
)

func (s *Service) Distribute(squads []*Squad, buildings []*Building) error {
	if len(squads) == 0 {
		return fmt.Errorf("no squads")
	}
	if len(buildings) == 0 {
		return fmt.Errorf("no buildings")
	}

	// 1. Копируем состояние для возможного отката
	buildingsCopy := deepCopyBuildings(buildings)

	// 2. Собираем всех мальчиков и девочек со всех отрядов
	var allBoys []*Participant
	var allGirls []*Participant

	for _, squad := range squads {
		for _, p := range squad.Participants {
			p.SquadID = squad.ID
			if p.Gender == "М" {
				allBoys = append(allBoys, p)
			} else {
				allGirls = append(allGirls, p)
			}
		}
	}

	// Сортируем: старшие идут первыми, так как им сложнее подобрать комнату по возрасту
	sortParticipantsByAgeDesc(allBoys)
	sortParticipantsByAgeDesc(allGirls)

	// =====================================================================
	// ФАЗА 1: Расселение по СТРОГО специализированным корпусам/этажам
	// =====================================================================

	// Селим мальчиков туда, где AllowedGender == "М"
	unplacedBoys := s.placeParticipants(allBoys, "М", buildingsCopy, true)
	// Селим девочек туда, где AllowedGender == "Ж"
	unplacedGirls := s.placeParticipants(allGirls, "Ж", buildingsCopy, true)

	// =====================================================================
	// ФАЗА 2: Расселение оставшихся в универсальные корпуса/этажи "М/Ж"
	// =====================================================================

	// Чтобы оптимизировать уплотнение, первой селим ту группу, которой осталось больше
	if len(unplacedGirls) > len(unplacedBoys) {
		unplacedGirls = s.placeParticipants(unplacedGirls, "Ж", buildingsCopy, false)
		unplacedBoys = s.placeParticipants(unplacedBoys, "М", buildingsCopy, false)
	} else {
		unplacedBoys = s.placeParticipants(unplacedBoys, "М", buildingsCopy, false)
		unplacedGirls = s.placeParticipants(unplacedGirls, "Ж", buildingsCopy, false)
	}

	// =====================================================================
	// ПРОВЕРКА РЕЗУЛЬТАТОВ
	// =====================================================================

	if len(unplacedBoys) > 0 {
		return fmt.Errorf("cannot distribute %d participants gender=М (например: %s, %d лет)",
			len(unplacedBoys), unplacedBoys[0].FullName, unplacedBoys[0].Age)
	}
	if len(unplacedGirls) > 0 {
		return fmt.Errorf("cannot distribute %d participants gender=Ж (например: %s, %d лет)",
			len(unplacedGirls), unplacedGirls[0].FullName, unplacedGirls[0].Age)
	}

	// Если всем хватило места, применяем изменения
	replaceBuildings(buildings, buildingsCopy)

	return nil
}

// placeParticipants пытается расселить список участников и возвращает тех, кому не хватило места
func (s *Service) placeParticipants(participants []*Participant, gender string, buildings []*Building, strict bool) []*Participant {
	var unplaced []*Participant

	for _, p := range participants {
		placed := false

		for _, b := range buildings {
			if placed {
				break
			}

			for _, f := range b.Floors {
				if placed {
					break
				}

				// Фильтр этажей: строгий режим (только свой пол) или М/Ж
				if strict {
					if f.AllowedGender != gender {
						continue
					}
				} else {
					if f.AllowedGender != "М/Ж" {
						continue
					}
				}

				// Сортируем комнаты по номеру для порядка (можно вынести, но для надежности оставим)
				sort.Slice(f.Rooms, func(i, j int) bool {
					return f.Rooms[i].Number < f.Rooms[j].Number
				})

				for _, r := range f.Rooms {
					// 1. Проверка вместимости
					if r.Occupied >= r.Capacity {
						continue
					}

					// 2. ПРОВЕРКА ПОЛА В КОМНАТЕ (Самое важное для М/Ж)
					// Если комната не пустая, проверяем пол первого заселенного
					if r.Occupied > 0 && r.Participants[0].Gender != gender {
						continue
					}

					// 3. Проверка возраста конкретного ребенка
					if p.Age >= r.MinAge && p.Age <= r.MaxAge {
						// Заселяем!
						r.Participants = append(r.Participants, p)
						r.Occupied++
						placed = true
						break // Прерываем цикл по комнатам, ребенок заселен
					}
				}
			}
		}

		// Если после обхода всех зданий ребенок не заселен, добавляем в список оставшихся
		if !placed {
			log.Printf("НЕ ЗАСЕЛЕН: %s, Пол: %s, Возраст: %d", p.FullName, p.Gender, p.Age)
			unplaced = append(unplaced, p)
		}
	}

	return unplaced
}

// Функции sortParticipantsByAgeDesc, deepCopyBuildings и replaceBuildings остаются без изменений!
func sortParticipantsByAgeDesc(participants []*Participant) {

	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Age >
			participants[j].Age
	})
}

func deepCopyBuildings(buildings []*Building) []*Building {

	result := make([]*Building, 0, len(buildings))

	for _, b := range buildings {

		nb := &Building{
			ID:          b.ID,
			Name:        b.Name,
			FloorsCount: b.FloorsCount,
			Capacity:    b.Capacity,
			Floors:      []*Floor{},
		}

		for _, f := range b.Floors {

			nf := &Floor{
				ID:               f.ID,
				Number:           f.Number,
				StartRoomsNumber: f.StartRoomsNumber,
				AllowedGender:    f.AllowedGender,
				CurrentGender:    f.CurrentGender,
				Rooms:            []*Room{},
			}

			for _, r := range f.Rooms {

				nr := &Room{
					ID:       r.ID,
					Number:   r.Number,
					MinAge:   r.MinAge,
					MaxAge:   r.MaxAge,
					Capacity: r.Capacity,
					Occupied: r.Occupied,
					Participants: append(
						[]*Participant{},
						r.Participants...,
					),
				}

				nf.Rooms = append(nf.Rooms, nr)
			}

			nb.Floors = append(nb.Floors, nf)
		}

		result = append(result, nb)
	}

	return result
}

func replaceBuildings(dst []*Building, src []*Building) {

	for i := range dst {

		for j := range dst[i].Floors {

			dst[i].Floors[j].CurrentGender =
				src[i].Floors[j].CurrentGender

			for k := range dst[i].Floors[j].Rooms {

				dst[i].Floors[j].Rooms[k].Occupied =
					src[i].Floors[j].Rooms[k].Occupied

				dst[i].Floors[j].Rooms[k].Participants =
					src[i].Floors[j].Rooms[k].Participants
			}
		}
	}
}
