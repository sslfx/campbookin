package api_gateway

import (
	pb2 "campbookin/api/proto/proto_output/admin_proto"
	pb1 "campbookin/api/proto/proto_output/reg_proto"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

type mockAuthClient struct {
	pb1.CampRegistrationClient
	LoginFunc func(ctx context.Context, in *pb1.LoginRequest) (*pb1.LoginResponse, error)
}

func (m *mockAuthClient) Login(ctx context.Context, in *pb1.LoginRequest, opts ...grpc.CallOption) (*pb1.LoginResponse, error) {
	return m.LoginFunc(ctx, in)
}

type mockAdminClient struct {
	pb2.CampAdminClient
	GetCampFunc   func(ctx context.Context, in *pb2.GetCampRequest) (*pb2.GetCampResponse, error)
	BookShiftFunc func(ctx context.Context, in *pb2.BookShiftRequest) (*pb2.BookShiftResponse, error)
}

func (m *mockAdminClient) GetCamp(ctx context.Context, in *pb2.GetCampRequest, opts ...grpc.CallOption) (*pb2.GetCampResponse, error) {
	if m.GetCampFunc == nil {
		return &pb2.GetCampResponse{}, nil
	}
	return m.GetCampFunc(ctx, in)
}

func (m *mockAdminClient) BookShift(ctx context.Context, in *pb2.BookShiftRequest, opts ...grpc.CallOption) (*pb2.BookShiftResponse, error) {
	if m.BookShiftFunc == nil {
		return &pb2.BookShiftResponse{Success: true}, nil
	}
	return m.BookShiftFunc(ctx, in)
}

func TestLogin(t *testing.T) {
	mockAuth := &mockAuthClient{
		LoginFunc: func(ctx context.Context, in *pb1.LoginRequest) (*pb1.LoginResponse, error) {
			return &pb1.LoginResponse{Token: "valid-token"}, nil
		},
	}
	gw := &Gateway{authClient: mockAuth}

	t.Run("success login", func(t *testing.T) {
		body := `{"email":"test@test.com", "password":"123"}`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
		w := httptest.NewRecorder()

		gw.Login(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "valid-token") {
			t.Error("token not found in response")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", strings.NewReader(`{bad json`))
		w := httptest.NewRecorder()
		gw.Login(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestGetCamp(t *testing.T) {
	mockAdmin := &mockAdminClient{
		GetCampFunc: func(ctx context.Context, in *pb2.GetCampRequest) (*pb2.GetCampResponse, error) {
			return &pb2.GetCampResponse{
				Camp: &pb2.Camp{Id: "1", Name: "Zarnitsa", Address: "Rzhev"},
			}, nil
		},
	}
	gw := &Gateway{adminClient: mockAdmin}

	req := httptest.NewRequest("GET", "/camps/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	gw.GetCamp(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBookShift_RateLimit(t *testing.T) {
	s := miniredis.RunT(t)
	rClient := redis.NewClient(&redis.Options{Addr: s.Addr()})

	mockAdmin := &mockAdminClient{
		BookShiftFunc: func(ctx context.Context, in *pb2.BookShiftRequest) (*pb2.BookShiftResponse, error) {
			return &pb2.BookShiftResponse{Success: true}, nil
		},
	}
	gw := &Gateway{adminClient: mockAdmin, redis: rClient}

	ctx := context.WithValue(context.Background(), userIDKey, "user123")

	req1 := httptest.NewRequest("POST", "/shifts/1/book", nil).WithContext(ctx)
	rctx1 := chi.NewRouteContext()
	rctx1.URLParams.Add("id", "1")
	req1 = req1.WithContext(context.WithValue(req1.Context(), chi.RouteCtxKey, rctx1))

	w1 := httptest.NewRecorder()
	gw.BookShift(w1, req1)

	req2 := httptest.NewRequest("POST", "/shifts/1/book", nil).WithContext(ctx)
	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("id", "1")
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))

	w2 := httptest.NewRecorder()
	gw.BookShift(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for second request, got %d", w2.Code)
	}
}
