package distribution

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

func (s *Service) CreatingSquads(kids []*Participant, squadsTotalCount int) ([]*Squad, error) {
	// 1. Сортируем ВСЕХ детей в едином списке от СТАРШИХ к МЛАДШИМ.
	// Так мы гарантируем, что самые взрослые пойдут в первые отряды вместе.
	sort.Slice(kids, func(i, j int) bool {
		return kids[i].Age > kids[j].Age // (по убыванию)
	})

	total := len(kids)
	squadCount := squadsTotalCount
	baseSize := total / squadCount
	remainder := total % squadCount

	squads := make([]*Squad, squadCount)
	for i := 0; i < squadCount; i++ {
		size := baseSize
		if i < remainder {
			size++
		}

		squads[i] = &Squad{
			ID:       strconv.Itoa(i + 1),
			Capacity: size,
		}
	}

	for _, k := range kids {
		if err := assign(k, squads); err != nil {
			return nil, err
		}
	}

	return squads, nil
}

func assign(p *Participant, squads []*Squad) error {
	var best *Squad
	bestScore := math.MaxFloat64
	pAge := float64(p.Age)

	// Добавляем индекс 'i' в цикл
	for i, s := range squads {
		if len(s.Participants) >= s.Capacity {
			continue // Отряд забит, идем дальше
		}

		score := 0.0

		// ПРИОРИТЕТ ОЧЕРЕДИ
		// Чем дальше отряд, тем больше штраф.
		score += float64(i) * 10.0

		// Возрастной штраф
		if len(s.Participants) > 0 {
			score += math.Abs(s.AvgAge-pAge) * 5.0
		}

		// Баланс пола
		fBoys, fGirls := s.BoysCount, s.GirlsCount
		if p.Gender == "М" {
			fBoys++
		} else {
			fGirls++
		}
		score += math.Abs(float64(fBoys-fGirls)) * 2.0

		if score < bestScore {
			bestScore = score
			best = s
		}
	}

	if best == nil {
		return fmt.Errorf("нет мест для ребенка: %s", p.FullName)
	}

	// Фиксируем данные в лучшем отряде
	best.Participants = append(best.Participants, p)
	p.SquadID = best.ID

	if p.Gender == "М" {
		best.BoysCount++
	} else {
		best.GirlsCount++
	}

	// Обновляем метрики отряда
	if len(best.Participants) == 1 {
		best.MinAge = pAge
		best.MaxAge = pAge
		best.AvgAge = pAge
	} else {
		if pAge < best.MinAge {
			best.MinAge = pAge
		}
		if pAge > best.MaxAge {
			best.MaxAge = pAge
		}

		// Скользящее среднее
		count := float64(len(best.Participants))
		best.AvgAge = (best.AvgAge*(count-1) + pAge) / count
	}

	return nil
}
