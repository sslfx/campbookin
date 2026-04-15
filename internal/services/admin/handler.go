package admin

import (
	pb "campbookin/api/proto/proto_output/admin_proto"
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedCampAdminServer
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  log,
	}
}

func (h *Handler) Assign(ctx context.Context, req *pb.AssignRequest) (*pb.AssignResponse, error) {

	h.logger.Info("AssignDirector: called")

	err := h.service.Assign(ctx, req.GetUserId(), req.GetCampId())
	if err != nil {
		h.logger.Error("AssignDirector error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AssignResponse{
		Success: true,
	}, nil
}

func (h *Handler) CreateCamp(ctx context.Context, req *pb.CreateCampRequest) (*pb.CreateCampResponse, error) {
	h.logger.Info("CreateCampHandler: GRPC CALLED")
	camp, err := h.service.CreateCamp(ctx, req.GetName(), req.GetDescription(), req.GetAddress())
	if err != nil {
		h.logger.Error("CreateCampHandler: service error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateCampResponse{Id: camp.ID}, nil
}

func (h *Handler) GetCamp(ctx context.Context, req *pb.GetCampRequest) (*pb.GetCampResponse, error) {
	h.logger.Info("GetCampHandler: GRPC CALLED")
	camp, err := h.service.GetCamp(ctx, req.GetId())
	if err != nil {
		h.logger.Error("GetCampHandler: service error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetCampResponse{
		Camp: &pb.Camp{
			Id:          camp.ID,
			Name:        camp.Name,
			Description: camp.Description,
			Address:     camp.Address,
		},
	}, nil
}

func (h *Handler) DeleteCamp(ctx context.Context, req *pb.DeleteCampRequest) (*pb.DeleteCampResponse, error) {
	h.logger.Info("DeleteCampHandler: GRPC CALLED")
	err := h.service.DeleteCamp(ctx, req.GetId())
	if err != nil {
		h.logger.Error("DeleteCampHandler: service error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteCampResponse{
		Success: true,
	}, nil
}

func (h *Handler) ListCamps(ctx context.Context, request *pb.ListCampsRequest) (*pb.ListCampsResponse, error) {
	h.logger.Info("ListCampsHandler: GRPC CALLED")
	camps, err := h.service.ListCamps(ctx)
	if err != nil {
		h.logger.Error("ListCampsHandler: service error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	campss := make([]*pb.Camp, len(camps))
	for i, camp := range camps {
		campss[i] = &pb.Camp{
			Id:          camp.ID,
			Name:        camp.Name,
			Description: camp.Description,
			Address:     camp.Address,
		}
	}
	return &pb.ListCampsResponse{
		Camps: campss,
	}, nil
}

func (h *Handler) UpdateCamp(ctx context.Context, req *pb.UpdateCampRequest) (*pb.UpdateCampResponse, error) {
	h.logger.Info("UpdateCampHandler: GRPC CALLED")
	err := h.service.UpdateCamp(ctx, req.GetId(), req.GetName(), req.GetDescription(), req.GetAddress())
	if err != nil {
		h.logger.Error("UpdateCampHandler: service error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateCampResponse{Success: true}, nil
}
