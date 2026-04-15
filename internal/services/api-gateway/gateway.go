package api_gateway

import (
	pb2 "campbookin/api/proto/proto_output/admin_proto"
	pb1 "campbookin/api/proto/proto_output/reg_proto"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Gateway struct {
	authClient  pb1.CampRegistrationClient
	adminClient pb2.CampAdminClient
	secret      []byte
}

func NewGateway(authClient pb1.CampRegistrationClient, adminClient pb2.CampAdminClient, secret []byte) *Gateway {
	return &Gateway{
		authClient:  authClient,
		adminClient: adminClient,
		secret:      secret,
	}
}
func (gw *Gateway) RegisterRoutes(r *chi.Mux) {
	r.Post("/registration", gw.Register)
	r.Post("/login", gw.Login)
	r.Group(func(protected chi.Router) {
		protected.Use(gw.JWTMiddleware)
		protected.Post("/camps", gw.CreateCamp)
		protected.Delete("/camps/{id}", gw.DeleteCamp)
		protected.Put("/camps/{id}", gw.UpdateCamp)
		protected.Get("/camps/{id}", gw.GetCamp)
		protected.Get("/camps", gw.ListCamps)
		protected.Post("/camps/{id}/assign", gw.AssignDirector)

	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	if err := json.NewEncoder(w).Encode(map[string]bool{"success": resp.Success}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
