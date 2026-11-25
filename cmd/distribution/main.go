package main

import (
	"campbookin/internal/services/distribution"
	"campbookin/pkg"
	"log"
	"os"
)

func main() {
	log.Println("Starting distribution")

	// 1. Инициализация
	if err := pkg.LoadEnv(".env"); err != nil {
		os.Exit(2)
	}

	s := distribution.NewService()

	// 1. Сначала получаем путь (строку)
	path := pkg.MustGet("PARSING_FILE")

	// 2. Открываем файл (получаем дескриптор file, который реализует io.Reader)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Не удалось открыть файл: ", err)
	}
	// Обязательно закрываем файл после использования
	defer file.Close()

	// 3. Теперь передаем этот file в функцию
	// (убедись, что метод в сервисе принимает io.Reader)
	p, err := s.ImportParticipants(file, "1210343434")
	if err != nil {
		log.Fatal("Ошибка импорта: ", err)
	}
	// 3. Создание отрядов
	squads, err := s.CreatingSquads(p, 8)
	if err != nil {
		log.Fatal("Ошибка создания отрядов: ", err)
	}

	// Экспорт отрядов (опционально)
	if err := s.ExportSquadsListForAdmins(squads, "/Users/max/Downloads/sq3.xlsx"); err != nil {
		log.Fatal("Ошибка экспорта отрядов: ", err)
	}
	log.Println("Отряды созданы и экспортированы")

	// 4. Подготовка строений (твоя логика)
	rooms1, _ := s.AddRooms(1, 8, 5, 10, 17)
	floor1, _ := s.CreateFloor(1, rooms1, "М/Ж")
	rooms2, _ := s.AddRooms(9, 8, 5, 6, 11)
	floor2, _ := s.CreateFloor(2, rooms2, "Ж")
	building1, _ := s.CreateBuilding("Кирпичный", 2, []*distribution.Floor{floor1, floor2})

	rooms3, _ := s.AddRooms(17, 4, 14, 6, 17)
	floor3, _ := s.CreateFloor(1, rooms3, "М")
	building2, _ := s.CreateBuilding("Столовая", 1, []*distribution.Floor{floor3})

	rooms4, _ := s.AddRooms(21, 8, 5, 6, 17)
	room30, _ := s.CreateRoom(30, 4, 0, 6, 17)
	rooms4 = append(rooms4, room30)
	floor4, _ := s.CreateFloor(1, rooms4, "М/Ж")
	building3, _ := s.CreateBuilding("Корпус 1", 1, []*distribution.Floor{floor4})

	rooms5, _ := s.AddRooms(31, 8, 5, 6, 17)
	room40, _ := s.CreateRoom(40, 4, 0, 6, 17)
	rooms5 = append(rooms5, room40)
	floor5, _ := s.CreateFloor(1, rooms5, "М/Ж")
	building4, _ := s.CreateBuilding("Корпус 2", 1, []*distribution.Floor{floor5})

	// 5. Распределение
	slice := []*distribution.Building{building1, building2, building3, building4}
	if err = s.Distribute(squads, slice); err != nil {
		log.Fatal("Ошибка распределения: ", err)
	}

	// 6. Финальный экспорт
	if err := s.ExportBuildingListForAdmins(slice, "/Users/max/Downloads/k.xlsx"); err != nil {
		log.Fatal("Ошибка экспорта зданий: ", err)
	}

	log.Println("Готово! Файл b.xlsx лежит в Downloads.")
}
