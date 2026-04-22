package notification

import (
	pb2 "campbookin/api/proto/proto_output/admin_proto"
	pb "campbookin/api/proto/proto_output/reg_proto"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"gopkg.in/gomail.v2"
)

type NotificationService struct {
	authClient  pb.CampRegistrationClient
	adminClient pb2.CampAdminClient
	redis       *redis.Client
}

func NewNotificationService(authClient pb.CampRegistrationClient, adminClient pb2.CampAdminClient, redisAddr string) *NotificationService {
	return &NotificationService{
		authClient:  authClient,
		adminClient: adminClient,
		redis: redis.NewClient(&redis.Options{
			Addr: redisAddr,
		}),
	}
}
func (n *NotificationService) StopRedis() error {
	log.Println("Shutting down redis listener")
	err := n.redis.Close()
	if err != nil {
		return err
	}
	return nil
}

type UserInfo struct {
	Name  string
	Email string
}
type ShiftStatus int

const (
	ShiftStatusUnknown  ShiftStatus = iota // 0
	ShiftStatusActive                      // 1
	ShiftStatusInactive                    // 2
)

type ShiftInfo struct {
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
type CampInfo struct {
	Name        string
	Description string
	Address     string
}

func (n *NotificationService) processBooking(ctx context.Context, email, pass string, event *BookingTicket) {
	log.Printf("Processing booking")
	userInfo, shiftInfo, campInfo, err := n.takeInformation(ctx, event.UserId, event.ShiftId)
	if err != nil {
		log.Printf("Error taking user information: %v", err)
		return
	}

	// 1. Генерируем PDF
	pdfPath, err := GeneratePDF(userInfo, shiftInfo, campInfo)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}
	defer func() { _ = os.Remove(pdfPath) }()
	err = n.SendEmailWithRetry(userInfo, pdfPath, pass, email)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return
	}

	log.Printf("Successfully sent voucher to %s", userInfo.Email)
}

func (n *NotificationService) takeInformation(ctx context.Context, userId, shiftId string) (*UserInfo, *ShiftInfo, *CampInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var shiftData *pb2.GetShiftResponse
	shiftKey := fmt.Sprintf("shift:%s", shiftId)

	err := n.getCachedData(ctx, shiftKey, &shiftData)
	if err != nil {
		log.Printf("Cache Miss for shift: %s", shiftId)
		resp, err := n.adminClient.GetShift(ctx, &pb2.GetShiftRequest{Id: shiftId})
		if err != nil {
			return nil, nil, nil, err
		}
		shiftData = resp
		n.setCache(ctx, shiftKey, &shiftData, 10*time.Minute)
	}

	var campData *pb2.GetCampResponse
	campKey := fmt.Sprintf("camp:%s", shiftData.Shift.CampId)

	err = n.getCachedData(ctx, campKey, &campData)
	if err != nil {
		log.Printf("Cache Miss for camp: %s", shiftData.Shift.CampId)
		resp, err := n.adminClient.GetCamp(ctx, &pb2.GetCampRequest{Id: shiftData.Shift.CampId})
		if err != nil {
			return nil, nil, nil, err
		}
		campData = resp
		n.setCache(ctx, campKey, &campData, 10*time.Minute)
	}

	var userData *pb.GetUserResponse
	userKey := fmt.Sprintf("user:%s", userId)

	err = n.getCachedData(ctx, userKey, &userData)
	if err != nil {
		log.Printf("Cache Miss for user: %s", userId)
		resp, err := n.authClient.GetUser(ctx, &pb.GetUserRequest{Id: userId})
		if err != nil {
			return nil, nil, nil, err
		}
		userData = resp
		n.setCache(ctx, userKey, &userData, 10*time.Minute)
	}

	return &UserInfo{
			Name:  userData.Name,
			Email: userData.Email,
		}, &ShiftInfo{
			Name:        shiftData.Shift.Name,
			Description: shiftData.Shift.Description,
			StartTime:   shiftData.Shift.StartTime.AsTime(),
			EndTime:     shiftData.Shift.EndTime.AsTime(),
			Capacity:    shiftData.Shift.Capacity,
			Price:       shiftData.Shift.Price,
			CampID:      shiftData.Shift.CampId,
			Status:      ShiftStatus(shiftData.Shift.Status),
		}, &CampInfo{
			Name:        campData.Camp.Name,
			Description: campData.Camp.Description,
			Address:     campData.Camp.Address,
		}, nil
}
func (n *NotificationService) getCachedData(ctx context.Context, key string, target interface{}) error {
	val, err := n.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), target)
}
func (n *NotificationService) setCache(ctx context.Context, key string, data interface{}, ttl time.Duration) {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal for cache: %v", err)
		return
	}
	n.redis.Set(ctx, key, bytes, ttl)
}

func GeneratePDF(user *UserInfo, shift *ShiftInfo, camp *CampInfo) (string, error) {
	// 1. Проверка наличия шрифта
	if _, err := os.Stat("Roboto.ttf"); os.IsNotExist(err) {
		log.Println("!!! CRITICAL: Roboto.ttf not found")
		return "", fmt.Errorf("font file missing")
	}

	// 2. Инициализация (P - портретная, mm - миллиметры, A4)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Добавляем UTF-8 шрифт. Мы НЕ используем транслятор ниже.
	pdf.AddUTF8Font("Roboto", "", "Roboto.ttf")
	pdf.SetFont("Roboto", "", 12)

	// --- ШАПКА ---
	pdf.SetTextColor(0, 51, 102) // Темно-синий
	pdf.SetFont("Roboto", "", 16)
	pdf.CellFormat(190, 15, "ЭЛЕКТРОННАЯ ПУТЕВКА", "0", 1, "C", false, 0, "")

	pdf.SetFont("Roboto", "", 10)
	pdf.CellFormat(190, 5, "Школа поискового движения «Наследники памяти»", "0", 1, "C", false, 0, "")
	pdf.Ln(10)

	// --- ЛИНИЯ-РАЗДЕЛИТЕЛЬ ---
	pdf.SetDrawColor(0, 51, 102)
	pdf.Line(10, 45, 200, 45)
	pdf.Ln(15)

	// --- ДАННЫЕ ПУТЕВКИ ---
	pdf.SetTextColor(0, 0, 0) // Возвращаем черный цвет

	// Вспомогательная функция, чтобы не дублировать код и проверять пустые значения
	addInfoRow := func(label, value string) {
		if value == "" {
			value = "Не указано"
		}
		pdf.SetFont("Roboto", "", 12)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(50, 10, label, "0", 0, "L", false, 0, "")

		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Roboto", "", 12)
		pdf.CellFormat(130, 10, value, "0", 1, "L", false, 0, "")
		pdf.Ln(2)
	}

	// Заполняем данными из твоего проекта
	addInfoRow("ФИО Родителя:", user.Name)
	addInfoRow("Название лагеря:", camp.Name)
	addInfoRow("Адрес лагеря:", camp.Address)
	addInfoRow("Название смены:", shift.Name)
	addInfoRow("Дата начала:", shift.StartTime.Format("02.01.2006 15:04"))
	addInfoRow("Дата окончания:", shift.EndTime.Format("02.01.2006 15:04"))

	pdf.Ln(10)

	// Блок описания
	pdf.SetFont("Roboto", "", 12)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(190, 10, "Описание программы:")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Roboto", "", 11)
	pdf.MultiCell(190, 6, shift.Description, "0", "L", false)

	// --- ПОДВАЛ ---
	pdf.SetY(-35)
	pdf.SetFont("Roboto", "", 9)
	pdf.SetTextColor(150, 150, 150)
	footerText := "Данная путевка подтверждает успешное бронирование в системе CampBookin. " +
		"Пожалуйста, предъявите её при заезде в лагерь"
	pdf.MultiCell(190, 5, footerText, "T", "C", false)

	// Сохранение
	fileName := fmt.Sprintf("ticket_%s.pdf", uuid.New().String())
	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании PDF: %w", err)
	}

	return fileName, nil
}
func (n *NotificationService) SendEmailWithRetry(user *UserInfo, path, pass, email string) error {
	var err error
	backoff := time.Second * 2

	for i := 0; i < 3; i++ {
		err = SendEmail(user, path, pass, email)
		if err == nil {
			return nil
		}

		log.Printf("Attempt %d failed to send email: %v. Retrying in %v...", i+1, err, backoff)
		time.Sleep(backoff)
		backoff *= 2
	}
	return fmt.Errorf("could not send email after 3 attempts: %w", err)
}
func SendEmail(user *UserInfo, path, pass, email string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", "Ваша путевка в лагерь!")
	m.SetBody("text/plain", "Поздравляем! Ваша путевка находится во вложении к этому письму.")

	m.Attach(path)

	d := gomail.NewDialer("smtp.gmail.com", 587, email, pass)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
