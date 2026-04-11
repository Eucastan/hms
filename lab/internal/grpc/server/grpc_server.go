package server

import (
	"context"
	"time"

	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/Eucastan/hms/lab/internal/services"
	labpb "github.com/Eucastan/hms/shared/pkg/proto/lab"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LabServiceServer struct {
	labpb.UnimplementedLabServiceServer
	labReqSvc    services.LabRequestService
	labResultSvc services.LabResultService
}

func NewLabServiceServer(labReq services.LabRequestService, labResult services.LabResultService) *LabServiceServer {
	return &LabServiceServer{labReqSvc: labReq, labResultSvc: labResult}
}

// Create LabRequest
func (s *LabServiceServer) CreateLabTestRequest(ctx context.Context, req *labpb.CreateLabRequest) (*labpb.LabResponse, error) {
	patientID, err := uuid.Parse(req.PatientId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	requestBy, err := uuid.Parse(req.RequestBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = s.labReqSvc.CreateLabRequest(ctx, patientID, requestBy, req.TestType, req.Priority, req.Status, req.Notes)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &labpb.LabResponse{Status: "created", Message: "Lab request created successfully"}, nil
}

// Get LabRequest
func (s *LabServiceServer) GetLabTestRequests(ctx context.Context, req *labpb.GetLabRequestsRequest) (*labpb.GetLabRequestsResponse, error) {
	patientID, err := uuid.Parse(req.PatientId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid patient id")
	}

	requests, err := s.labReqSvc.GetLabRequestsByPatient(ctx, patientID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve lab requests")
	}

	var protoReqs []*labpb.LabTestRequest
	for _, r := range requests {
		protoReqs = append(protoReqs, &labpb.LabTestRequest{
			Id:        r.ID.String(),
			PatientId: r.PatientID.String(),
			RequestBy: r.RequestBy.String(),
			TestType:  r.TestType,
			Priority:  r.Priority,
			Status:    r.Status,
			Notes:     r.Notes,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}

	return &labpb.GetLabRequestsResponse{Requests: protoReqs}, nil
}

// Update LabRequest
func (s *LabServiceServer) UpdateLabTestRequest(ctx context.Context, req *labpb.UpdateLabRequest) (*labpb.LabResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	requestBy, err := uuid.Parse(req.RequestBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	labReq := &models.LabRequestUpdate{
		TestType: &req.TestType,
		Priority: &req.Priority,
		Notes:    &req.Notes,
	}

	err = s.labReqSvc.UpdateLabRequest(ctx, id, requestBy, labReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &labpb.LabResponse{
		Status:  "updated",
		Message: "Lab request updated successfully",
	}, nil
}

// Delete LabRequest
func (s *LabServiceServer) DeleteLabTestRequest(ctx context.Context, req *labpb.DeleteLabRequest) (*labpb.LabResponse, error) {
	id := uuid.MustParse(req.Id)
	err := s.labReqSvc.DeleteLabRequest(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &labpb.LabResponse{Status: "deleted", Message: "Lab request deleted successfully"}, nil
}

// LabResult handlers
func (s *LabServiceServer) GetLabTestResult(ctx context.Context, req *labpb.GetLabResultRequest) (*labpb.LabResultResponse, error) {

	patientID := uuid.MustParse(req.PatientId)

	results, err := s.labResultSvc.GetByPatientID(ctx, patientID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoResults []*labpb.LabResult
	for _, r := range results {
		protoResults = append(protoResults, &labpb.LabResult{
			PatientID:      r.PatientID.String(),
			PerformedBy:    r.PerformedBy.String(),
			TestType:       r.TestType,
			ResultValue:    r.ResultValue,
			Unit:           r.Unit,
			ReferenceRange: r.ReferenceRange,
			Comments:       r.Comments,
			Verified:       r.Verified,
		})
	}

	return &labpb.LabResultResponse{Results: protoResults}, nil
}
