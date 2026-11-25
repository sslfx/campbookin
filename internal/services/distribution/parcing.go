package distribution

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var phoneRegex = regexp.MustCompile(`(\+7|8|7)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}`)

func (s *Service) ImportParticipants(r io.Reader, shiftID string) ([]*Participant, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}

	// Получи список всех листов
	sheets := f.GetSheetList()
	log.Printf("Листы в файле: %v", sheets)

	if len(sheets) == 0 {
		return nil, fmt.Errorf("в файле нет ни одного листа")
	}

	// Попробуй явно указать имя первого листа вместо пустого или индекса
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	var participants []*Participant
	dataStarted := false

	var lastOrg string

	for _, row := range rows {
		// 1. начало данных. это после "№ п/п"
		if !dataStarted {
			if len(row) > 1 && (row[0] == "№ п/п" || row[1] == "Ф.И.О.") {
				dataStarted = true
			}
			continue
		}

		// 2. Пропускаем пустые строки
		if len(row) < 8 || row[1] == "" {
			continue
		}

		currentOrg := row[2]
		// Если ячейка пустая, значит она была частью объединения
		if currentOrg == "" {
			currentOrg = lastOrg // Наследуем значение из предыдущей строки
		} else {
			lastOrg = currentOrg // Запоминаем новое значение для следующих пустых
		}
		// 3. Парсим дату рождения
		birthDate, err := time.Parse("02.01.2006", strings.TrimSpace(row[4]))
		if err != nil {
			log.Fatal(err)
		}
		now := time.Now()
		age := now.Year() - birthDate.Year()
		if now.Before(birthDate.AddDate(age, 0, 0)) {
			age--
		}
		if age < 6 || age > 18 {
			log.Println("age out of range ", row[1])
			return nil, errors.New("age out of range")
		}
		// 4. Разделяем родителя и телефон
		pName, pPhone := splitParentData(row[6])

		p := &Participant{
			FullName:     strings.TrimSpace(row[1]),
			Organization: lastOrg,
			BirthDate:    birthDate,
			Age:          age,
			ParentName:   pName,
			ParentPhone:  pPhone,
			Gender:       strings.TrimSpace(row[7]),
			ShiftID:      shiftID,
		}
		participants = append(participants, p)
	}

	return participants, nil
}

func splitParentData(raw string) (name, phone string) {
	phone = phoneRegex.FindString(raw)
	name = phoneRegex.ReplaceAllString(raw, "")
	name = strings.TrimSpace(name)
	// Убираем лишние пробелы внутри имени, если они остались после удаления телефона
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	return name, phone
}
