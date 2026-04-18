package notification

import (
	pb2 "campbookin/api/proto/proto_output/admin_proto"
	pb "campbookin/api/proto/proto_output/reg_proto"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"gopkg.in/gomail.v2"
)

type NotificationService struct {
	authClient  pb.CampRegistrationClient
	adminClient pb2.CampAdminClient
}

func NewNotificationService(authClient pb.CampRegistrationClient, adminClient pb2.CampAdminClient) *NotificationService {
	return &NotificationService{
		authClient:  authClient,
		adminClient: adminClient,
	}
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

func (n *NotificationService) takeInformation(ctx context.Context, userId, shiftId string) (*UserInfo, *ShiftInfo, *CampInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	shift, err := n.adminClient.GetShift(ctx, &pb2.GetShiftRequest{Id: shiftId})
	if err != nil {
		return nil, nil, nil, err
	}
	camp, err := n.adminClient.GetCamp(ctx, &pb2.GetCampRequest{Id: shift.Shift.CampId})
	if err != nil {
		return nil, nil, nil, err
	}
	user, err := n.authClient.GetUser(ctx, &pb.GetUserRequest{Id: userId})
	if err != nil {
		return nil, nil, nil, err
	}
	return &UserInfo{
			Name:  user.Name,
			Email: user.Email,
		}, &ShiftInfo{
			Name:        shift.Shift.Name,
			Description: shift.Shift.Description,
			StartTime:   shift.Shift.StartTime.AsTime(),
			EndTime:     shift.Shift.EndTime.AsTime(),
			Capacity:    shift.Shift.Capacity,
			Price:       shift.Shift.Price,
			CampID:      shift.Shift.CampId,
			Status:      ShiftStatus(shift.Shift.Status),
		}, &CampInfo{
			Name:        camp.Camp.Name,
			Description: camp.Camp.Description,
			Address:     camp.Camp.Address,
		}, nil
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
	defer os.Remove(pdfPath)
	err = SendEmail(userInfo, pdfPath, pass, email)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return
	}

	log.Printf("Successfully sent voucher to %s", userInfo.Email)
}

func GeneratePDF(user *UserInfo, shift *ShiftInfo, camp *CampInfo) (string, error) {
	if _, err := os.Stat("Roboto.ttf"); os.IsNotExist(err) {
		log.Println("!!! CRITICAL: Roboto.ttf not found in working directory")
		return "", fmt.Errorf("font file missing")
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("Roboto", "", "Roboto.ttf")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	//header
	pdf.SetFont("Roboto", "", 12)
	pdf.SetTextColor(0, 51, 102)
	pdf.CellFormat(190, 15, tr("ЭЛЕКТРОННАЯ ПУТЕВКА"), "0", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Данные путевки
	pdf.SetFont("Roboto", "", 14)
	pdf.SetTextColor(0, 0, 0) // Черный

	// Рисуем поля данных
	data := [][]string{

		{"ФИО Родителя:", user.Name},
		{"Название лагеря:", camp.Name},
		{"Адресс лагеря:", camp.Address},
		{"Название смены:", shift.Name},
		{"Описание смены:", shift.Description},
		{"Дата начала:", shift.StartTime.Format("2006-01-02 15:04:05")},
		{"Дата окончания:", shift.EndTime.Format("2006-01-02 15:04:05")},
	}

	for _, row := range data {
		pdf.SetFont("Roboto", "", 12)
		pdf.CellFormat(50, 10, tr(row[0]), "0", 0, "L", false, 0, "")

		pdf.SetFont("Roboto", "", 12)
		pdf.CellFormat(130, 10, tr(row[1]), "0", 1, "L", false, 0, "")
		pdf.Ln(2)
	}

	pdf.Ln(20)
	// Подвал
	pdf.SetFont("Roboto", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.MultiCell(190, 5, tr("Пожалуйста, распечатайте данную путевку и предъявите её при заезде в лагерь вместе с медицинскими справками."), "0", "C", false)

	// Путь сохранения

	fileName := fmt.Sprintf("ticket_%s.pdf", uuid.New().String())

	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании PDF: %w", err)
	}

	return fileName, nil
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
