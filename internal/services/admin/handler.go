package admin

import (
	pb "campbookin/api/proto/proto_output/admin_proto"
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		h.logger.Error("AssignDirector error", "error", err.Error())
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
		h.logger.Error("CreateCampHandler: service error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateCampResponse{Id: camp.ID}, nil
}

func (h *Handler) GetCamp(ctx context.Context, req *pb.GetCampRequest) (*pb.GetCampResponse, error) {
	h.logger.Info("GetCampHandler: GRPC CALLED")
	camp, err := h.service.GetCamp(ctx, req.GetId())
	if err != nil {
		h.logger.Error("GetCampHandler: service error", "error", err.Error())
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
		h.logger.Error("DeleteCampHandler: service error", "error", err.Error())
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
		h.logger.Error("ListCampsHandler: service error", "error", err.Error())
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
		h.logger.Error("UpdateCampHandler: service error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateCampResponse{Success: true}, nil
}

func (h *Handler) CreateShift(ctx context.Context, req *pb.CreateShiftRequest) (*pb.CreateShiftResponse, error) {
	h.logger.Info("CreateShiftHandler: GRPC CALLED")
	shift, err := h.service.CreateShift(ctx, req.GetName(), req.GetCampId(), req.GetDescription(), req.GetCapacity(), req.GetPrice(), ShiftStatus(req.GetStatus()), req.GetStartTime().AsTime(), req.GetEndTime().AsTime())
	if err != nil {
		h.logger.Error("CreateShiftHandler: service error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateShiftResponse{Id: shift.ID}, nil
}
func mapStatusToProto(s ShiftStatus) pb.Status {
	switch s {
	case ShiftStatusActive:
		return pb.Status_SHIFT_ACTIVE
	case ShiftStatusInactive:
		return pb.Status_SHIFT_INACTIVE
	default:
		return pb.Status_SHIFT_UNKNOWN
	}
}
func (h *Handler) GetShift(ctx context.Context, req *pb.GetShiftRequest) (*pb.GetShiftResponse, error) {
	h.logger.Info("GetShiftHandler: GRPC CALLED")
	shift, err := h.service.GetShift(ctx, req.GetId())
	if err != nil {
		h.logger.Error("GetShiftHandler: service error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetShiftResponse{
		Shift: &pb.Shift{
			Id:          shift.ID,
			Name:        shift.Name,
			CampId:      shift.CampID,
			Description: shift.Description,
			Status:      mapStatusToProto(shift.Status),
			Price:       shift.Price,
			Capacity:    shift.Capacity,
			EndTime:     timestamppb.New(shift.EndTime),
			StartTime:   timestamppb.New(shift.StartTime),
		},
	}, nil
}

func (h *Handler) UpdateShift(ctx context.Context, req *pb.UpdateShiftRequest) (*pb.UpdateShiftResponse, error) {
	h.logger.Info("UpdateShiftHandler: GRPC CALLED")
	err := h.service.UpdateShift(ctx, req.GetId(), req.GetName(), req.GetDescription(), req.GetStartTime().AsTime(), req.GetEndTime().AsTime(), req.GetCapacity(), req.GetPrice(), ShiftStatus(req.GetStatus()))
	if err != nil {
		h.logger.Error("UpdateShiftHandler: service error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateShiftResponse{Success: true}, nil
}

func (h *Handler) DeleteShift(ctx context.Context, req *pb.DeleteShiftRequest) (*pb.DeleteShiftResponse, error) {
	h.logger.Info("DeleteShiftHandler: GRPC CALLED")
	err := h.service.DeleteShift(ctx, req.GetId())
	if err != nil {
		h.logger.Error("DeleteShiftHandler: service error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteShiftResponse{Success: true}, nil
}

func (h *Handler) ListShifts(ctx context.Context, req *pb.ListShiftsRequest) (*pb.ListShiftsResponse, error) {
	h.logger.Info("ListShiftsHandler: GRPC CALLED")
	shifts, err := h.service.ListShifts(ctx)
	if err != nil {
		h.logger.Error("ListShiftsHandler: service error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	shiftsResp := make([]*pb.Shift, len(shifts))
	for i, s := range shifts {
		shiftsResp[i] = &pb.Shift{
			Id:          s.ID,
			Name:        s.Name,
			CampId:      s.CampID,
			Description: s.Description,
			Status:      mapStatusToProto(s.Status),
			Price:       s.Price,
			Capacity:    s.Capacity,
			EndTime:     timestamppb.New(s.EndTime),
			StartTime:   timestamppb.New(s.StartTime),
		}
	}
	return &pb.ListShiftsResponse{
		Shifts: shiftsResp,
	}, nil
}

func (h *Handler) ListShiftsForCamp(ctx context.Context, req *pb.ListShiftsForCampRequest) (*pb.ListShiftsForCampResponse, error) {
	h.logger.Info("ListShiftsForCampHandler: GRPC CALLED")
	shifts, err := h.service.ListShiftsForCamp(ctx, req.GetCampId())
	if err != nil {
		h.logger.Error("ListShiftsForCampHandler: service error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	shiftsResp := make([]*pb.Shift, len(shifts))
	for i, s := range shifts {
		shiftsResp[i] = &pb.Shift{
			Id:          s.ID,
			Name:        s.Name,
			CampId:      s.CampID,
			Description: s.Description,
			Status:      mapStatusToProto(s.Status),
			Price:       s.Price,
			Capacity:    s.Capacity,
			EndTime:     timestamppb.New(s.EndTime),
			StartTime:   timestamppb.New(s.StartTime),
		}
	}
	return &pb.ListShiftsForCampResponse{
		Shifts: shiftsResp,
	}, nil
}

func (h *Handler) BookShift(ctx context.Context, req *pb.BookShiftRequest) (*pb.BookShiftResponse, error) {
	h.logger.Info("BookingShiftHandler: GRPC CALLED")
	success, err := h.service.BookShift(ctx, req.GetShiftId(), req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !success {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.BookShiftResponse{Success: success}, nil
}
