package distribution

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func (s *Service) ExportSquadsListForAdmins(squads []*Squad, filepath string) error {
	f := excelize.NewFile()
	defer f.Close()

	for i, squad := range squads {
		sheetName := "Squad" + squad.ID

		if i == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {

			_, err := f.NewSheet(sheetName)
			if err != nil {
				return err
			}
		}
		headers := []string{"ФИО", "Пол", "Возраст", "Организация", "Родитель", "Телефон"}
		for col, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		})
		f.SetCellStyle(sheetName, "A1", "F1", style)

		for rowIdx, p := range squad.Participants {
			row := rowIdx + 2 // Начинаем со второй строки (после заголовка)
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), p.FullName)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), p.Gender)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), p.Age)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), p.Organization)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), p.ParentName)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), p.ParentPhone)
		}

		// 4. Добавляем итоговую инфу по отряду внизу
		lastRow := len(squad.Participants) + 4
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", lastRow), "ИТОГО:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", lastRow), fmt.Sprintf("Мальчиков: %d, Девочек: %d", squad.BoysCount, squad.GirlsCount))
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", lastRow+1), "Средний возраст:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", lastRow+1), fmt.Sprintf("%.2f", squad.AvgAge))
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", lastRow+2), "Общее количество детей в отряде:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", lastRow+2), fmt.Sprintf("%d", squad.BoysCount+squad.GirlsCount))

		// Авто-подбор ширины колонок (простой вариант)
		f.SetColWidth(sheetName, "A", "A", 30)
		f.SetColWidth(sheetName, "D", "F", 20)

	}

	if err := f.SaveAs(filepath); err != nil {
		return err
	}

	return nil
}

func (s *Service) ExportBuildingListForAdmins(buildings []*Building, filepath string) error {
	f := excelize.NewFile()
	defer f.Close()

	for i, building := range buildings {
		// Название листа — имя корпуса
		sheetName := building.Name
		if sheetName == "" {
			sheetName = fmt.Sprintf("Корпус %d", i+1)
		}

		if i == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			_, err := f.NewSheet(sheetName)
			if err != nil {
				return err
			}
		}

		// Заголовки (добавили Этаж и Номер комнаты)
		headers := []string{"Этаж", "Комната", "ФИО", "Пол", "Возраст", "Отряд", "Родитель", "Телефон"}
		for col, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Стиль заголовков
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font:   &excelize.Font{Bold: true},
			Fill:   excelize.Fill{Type: "pattern", Color: []string{"#C5D9F1"}, Pattern: 1},
			Border: []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
		})
		f.SetCellStyle(sheetName, "A1", "H1", headerStyle)

		currentRow := 2
		totalOccupied := 0

		// Идем по иерархии: Здание -> Этаж -> Комната -> Участники
		for _, floor := range building.Floors {
			for _, room := range floor.Rooms {
				for _, p := range room.Participants {
					f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), floor.Number)
					f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), room.Number)
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), p.FullName)
					f.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), p.Gender)
					f.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), p.Age)
					f.SetCellValue(sheetName, fmt.Sprintf("F%d", currentRow), p.SquadID)
					f.SetCellValue(sheetName, fmt.Sprintf("G%d", currentRow), p.ParentName)
					f.SetCellValue(sheetName, fmt.Sprintf("H%d", currentRow), p.ParentPhone)

					currentRow++
					totalOccupied++
				}
			}
		}

		// 4. Итоговая информация по корпусу
		summaryRow := currentRow + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ИТОГО ПО КОРПУСУ:")
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), fmt.Sprintf("Заселено: %d из %d", totalOccupied, building.Capacity))

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow+1), "Свободных мест:")
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow+1), building.Capacity-totalOccupied)

		// Настройка ширины колонок
		f.SetColWidth(sheetName, "A", "B", 10) // Этаж и комната
		f.SetColWidth(sheetName, "C", "C", 35) // ФИО
		f.SetColWidth(sheetName, "D", "F", 12) // Пол, Возраст, Отряд
		f.SetColWidth(sheetName, "G", "H", 25) // Родители и телефон
	}

	if err := f.SaveAs(filepath); err != nil {
		return err
	}

	return nil
}
