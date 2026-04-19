package api_gateway

import (
	pb2 "campbookin/api/proto/proto_output/admin_proto"
	pb1 "campbookin/api/proto/proto_output/reg_proto"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Gateway struct {
	authClient  pb1.CampRegistrationClient
	adminClient pb2.CampAdminClient
	secret      []byte
	redis       *redis.Client
}

func NewGateway(authClient pb1.CampRegistrationClient, adminClient pb2.CampAdminClient, secret []byte, addr string) *Gateway {
	return &Gateway{
		authClient:  authClient,
		adminClient: adminClient,
		secret:      secret,
		redis: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		}),
	}
}
func (gw *Gateway) RegisterRoutes(r *chi.Mux) {
	r.Post("/registration", gw.Register)
	r.Post("/login", gw.Login)
	r.Group(func(protected chi.Router) {
		protected.Use(gw.JWTMiddleware)
		protected.Use(gw.RateLimiterMiddleware)
		protected.Post("/camps", gw.CreateCamp)
		protected.Delete("/camps/{id}", gw.DeleteCamp)
		protected.Put("/camps/{id}", gw.UpdateCamp)
		protected.Get("/camps/{id}", gw.GetCamp)
		protected.Get("/camps/{id}/shifts", gw.ListShiftsForCamp)
		protected.Get("/camps", gw.ListCamps)
		protected.Post("/camps/{id}/assign", gw.AssignDirector)
		protected.Post("/shifts", gw.CreateShift)
		protected.Delete("/shifts/{id}", gw.DeleteShift)
		protected.Put("/shifts/{id}", gw.UpdateShift)
		protected.Get("/shifts/{id}", gw.GetShift)
		protected.Get("/shifts", gw.ListShifts)
		protected.Post("/shifts/{id}/book", gw.BookShift)
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}
type LoginResponse struct {
	Token string `json:"token"`
}
type RegisterResponse struct {
	ID string `json:"id"`
}

func (gw *Gateway) Login(w http.ResponseWriter, req *http.Request) {
	var loginRequest LoginRequest
	err := json.NewDecoder(req.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	loginResp, err := gw.authClient.Login(req.Context(), &pb1.LoginRequest{
		Email:    loginRequest.Email,
		Password: loginRequest.Password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(LoginResponse{
		Token: loginResp.Token,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (gw *Gateway) Register(w http.ResponseWriter, req *http.Request) {
	var regReq LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&regReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	regResp, err := gw.authClient.Registration(req.Context(), &pb1.RegistrationRequest{
		Email:    regReq.Email,
		Password: regReq.Password,
		Name:     regReq.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(RegisterResponse{
		ID: regResp.Id,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type CreateCampRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
}
type CreateCampResponse struct {
	Id string `json:"id"`
}

func (gw *Gateway) CreateCamp(w http.ResponseWriter, req *http.Request) {
	fmt.Println("--> CreateCamp Handler Started")
	var createCampReq CreateCampRequest
	if err := json.NewDecoder(req.Body).Decode(&createCampReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.CreateCamp(ctx, &pb2.CreateCampRequest{
		Name:        createCampReq.Name,
		Description: createCampReq.Description,
		Address:     createCampReq.Address,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(CreateCampResponse{
		Id: resp.Id,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type UpdateCampRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
}
type UpdateCampResponse struct {
	Success bool `json:"success"`
}

func (gw *Gateway) UpdateCamp(w http.ResponseWriter, req *http.Request) {
	campID := chi.URLParam(req, "id")
	if campID == "" {
		http.Error(w, "missing camp id", http.StatusBadRequest)
		return
	}
	var updateCampReq UpdateCampRequest
	if err := json.NewDecoder(req.Body).Decode(&updateCampReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.UpdateCamp(ctx, &pb2.UpdateCampRequest{
		Id:          campID,
		Name:        updateCampReq.Name,
		Description: updateCampReq.Description,
		Address:     updateCampReq.Address,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(UpdateCampResponse{
		Success: resp.Success,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type DeleteCampResponse struct {
	Success bool `json:"success"`
}

func (gw *Gateway) DeleteCamp(w http.ResponseWriter, req *http.Request) {
	campID := chi.URLParam(req, "id")
	if campID == "" {
		http.Error(w, "missing camp id", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.DeleteCamp(ctx, &pb2.DeleteCampRequest{
		Id: campID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(DeleteCampResponse{
		Success: resp.Success,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type GetCampResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
}

func (gw *Gateway) GetCamp(w http.ResponseWriter, req *http.Request) {
	campID := chi.URLParam(req, "id")
	if campID == "" {
		http.Error(w, "missing camp id", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.GetCamp(ctx, &pb2.GetCampRequest{
		Id: campID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := GetCampResponse{
		Id:          resp.Camp.Id,
		Name:        resp.Camp.Name,
		Description: resp.Camp.Description,
		Address:     resp.Camp.Address,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type ListCampsResponse struct {
	Camps []GetCampResponse `json:"camps"`
}

func (gw *Gateway) ListCamps(w http.ResponseWriter, req *http.Request) {
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.ListCamps(ctx, &pb2.ListCampsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var camps []GetCampResponse
	for _, camp := range resp.Camps {
		camps = append(camps, GetCampResponse{
			Id:          camp.Id,
			Name:        camp.Name,
			Description: camp.Description,
			Address:     camp.Address,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ListCampsResponse{
		Camps: camps,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

type AssignDirectorRequest struct {
	UserID string `json:"user_id"`
}
type AssignDirectorResponse struct {
	Success bool `json:"success"`
}

func (gw *Gateway) AssignDirector(w http.ResponseWriter, req *http.Request) {
	campID := chi.URLParam(req, "id")
	if campID == "" {
		http.Error(w, "missing camp id", http.StatusBadRequest)
		return
	}
	var assignDirectorReq AssignDirectorRequest
	if err := json.NewDecoder(req.Body).Decode(&assignDirectorReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.Assign(ctx, &pb2.AssignRequest{
		UserId: assignDirectorReq.UserID,
		CampId: campID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(AssignDirectorResponse{Success: resp.Success}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type BookShiftResponse struct {
	Success bool `json:"success"`
}

func (gw *Gateway) BookShift(w http.ResponseWriter, req *http.Request) {
	shiftID := chi.URLParam(req, "id")
	ctx := gw.withAuthContext(req)
	userID, _ := req.Context().Value("userID").(string)

	rateKey := fmt.Sprintf("rate:book:%s:%s", userID, shiftID)
	val, err := gw.redis.Incr(ctx, rateKey).Result()
	if err == nil && val == 1 {
		gw.redis.Expire(ctx, rateKey, 5*time.Second)
	}
	if val > 1 {
		http.Error(w, "Слишком много запросов. Подождите немного.", http.StatusTooManyRequests)
		return
	}

	resp, err := gw.adminClient.BookShift(ctx, &pb2.BookShiftRequest{
		ShiftId: shiftID,
		UserId:  userID,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(BookShiftResponse{Success: resp.Success}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type ShiftStatus int

const (
	ShiftStatusUnknown  ShiftStatus = iota // 0
	ShiftStatusActive                      // 1
	ShiftStatusInactive                    // 2
)

type CreateShiftRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
	Capacity    int32       `json:"capacity"`
	Price       int32       `json:"price"`
	CampID      string      `json:"camp_id"`
	Status      ShiftStatus `json:"status"`
}
type CreateShiftResponse struct {
	Id string `json:"id"`
}

func (gw *Gateway) CreateShift(w http.ResponseWriter, req *http.Request) {
	fmt.Println("CreateCamp Handler Started")
	var createShiftReq CreateShiftRequest
	if err := json.NewDecoder(req.Body).Decode(&createShiftReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.CreateShift(ctx, &pb2.CreateShiftRequest{
		Name:        createShiftReq.Name,
		Description: createShiftReq.Description,
		StartTime:   timestamppb.New(createShiftReq.StartTime),
		EndTime:     timestamppb.New(createShiftReq.EndTime),
		Capacity:    createShiftReq.Capacity,
		CampId:      createShiftReq.CampID,
		Status:      pb2.Status(createShiftReq.Status),
		Price:       createShiftReq.Price,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(CreateShiftResponse{
		Id: resp.Id,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type GetShiftResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
	Capacity    int32       `json:"capacity"`
	Price       int32       `json:"price"`
	CampID      string      `json:"camp_id"`
	Status      ShiftStatus `json:"status"`
}

func (gw *Gateway) GetShift(w http.ResponseWriter, req *http.Request) {
	shiftID := chi.URLParam(req, "id")
	if shiftID == "" {
		http.Error(w, "missing shift id", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.GetShift(ctx, &pb2.GetShiftRequest{
		Id: shiftID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := GetShiftResponse{
		ID:          resp.Shift.Id,
		Name:        resp.Shift.Name,
		Description: resp.Shift.Description,
		StartTime:   resp.Shift.StartTime.AsTime(),
		EndTime:     resp.Shift.EndTime.AsTime(),
		Capacity:    resp.Shift.Capacity,
		Price:       resp.Shift.Price,
		CampID:      resp.Shift.CampId,
		Status:      ShiftStatus(resp.Shift.Status),
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type UpdateShiftRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
	Capacity    int32       `json:"capacity"`
	Price       int32       `json:"price"`
	CampID      string      `json:"camp_id"`
	Status      ShiftStatus `json:"status"`
}
type UpdateShiftResponse struct {
	Success bool `json:"success"`
}

func mapStatusToProto(s ShiftStatus) pb2.Status {
	switch s {
	case ShiftStatusActive:
		return pb2.Status_SHIFT_ACTIVE
	case ShiftStatusInactive:
		return pb2.Status_SHIFT_INACTIVE
	default:
		return pb2.Status_SHIFT_UNKNOWN
	}
}
func (gw *Gateway) UpdateShift(w http.ResponseWriter, req *http.Request) {
	shiftID := chi.URLParam(req, "id")
	if shiftID == "" {
		http.Error(w, "missing shift id", http.StatusBadRequest)
		return
	}
	var updateShiftReq UpdateShiftRequest
	if err := json.NewDecoder(req.Body).Decode(&updateShiftReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.UpdateShift(ctx, &pb2.UpdateShiftRequest{
		Id:          shiftID,
		Name:        updateShiftReq.Name,
		Description: updateShiftReq.Description,
		StartTime:   timestamppb.New(updateShiftReq.StartTime),
		EndTime:     timestamppb.New(updateShiftReq.EndTime),
		Capacity:    updateShiftReq.Capacity,
		Price:       updateShiftReq.Price,
		CampId:      updateShiftReq.CampID,
		Status:      mapStatusToProto(updateShiftReq.Status),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(UpdateShiftResponse{
		Success: resp.Success,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type DeleteShiftResponse struct {
	Success bool `json:"success"`
}

func (gw *Gateway) DeleteShift(w http.ResponseWriter, req *http.Request) {
	shiftId := chi.URLParam(req, "id")
	if shiftId == "" {
		http.Error(w, "missing shift id", http.StatusBadRequest)
		return
	}
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.DeleteShift(ctx, &pb2.DeleteShiftRequest{
		Id: shiftId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(DeleteShiftResponse{
		Success: resp.Success,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type ListShiftsResponse struct {
	Shifts []GetShiftResponse `json:"shifts"`
}

func (gw *Gateway) ListShifts(w http.ResponseWriter, req *http.Request) {
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.ListShifts(ctx, &pb2.ListShiftsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var shifts []GetShiftResponse
	for _, shift := range resp.Shifts {
		shifts = append(shifts, GetShiftResponse{
			ID:          shift.Id,
			Name:        shift.Name,
			Description: shift.Description,
			StartTime:   shift.StartTime.AsTime(),
			EndTime:     shift.EndTime.AsTime(),
			Capacity:    shift.Capacity,
			Price:       shift.Price,
			CampID:      shift.CampId,
			Status:      ShiftStatus(shift.Status),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ListShiftsResponse{
		Shifts: shifts,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
func (gw *Gateway) ListShiftsForCamp(w http.ResponseWriter, req *http.Request) {
	campID := chi.URLParam(req, "id")
	ctx := gw.withAuthContext(req)
	resp, err := gw.adminClient.ListShiftsForCamp(ctx, &pb2.ListShiftsForCampRequest{
		CampId: campID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var shifts []GetShiftResponse
	for _, shift := range resp.Shifts {
		shifts = append(shifts, GetShiftResponse{
			ID:          shift.Id,
			Name:        shift.Name,
			Description: shift.Description,
			StartTime:   shift.StartTime.AsTime(),
			EndTime:     shift.EndTime.AsTime(),
			Capacity:    shift.Capacity,
			Price:       shift.Price,
			CampID:      shift.CampId,
			Status:      ShiftStatus(shift.Status),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ListShiftsResponse{
		Shifts: shifts,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
