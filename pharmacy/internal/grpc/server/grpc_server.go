package server

import (
	"context"

	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/Eucastan/hms/pharmacy/internal/services"
	pharm "github.com/Eucastan/hms/shared/pkg/proto/pharmacy"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PharmacyServer struct {
	pharm.UnimplementedPharmacyServiceServer
	dispenseSvc services.DispenseService
}

func NewPharmacyServer(dispense services.DispenseService) *PharmacyServer {
	return &PharmacyServer{dispenseSvc: dispense}
}

func (s *PharmacyServer) CreatePrescription(ctx context.Context, req *pharm.PrescriptionRequest) (*pharm.PrescriptionResponse, error) {
	doctorID, err := uuid.Parse(req.DoctorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid doctor id")
	}

	input := &models.CreatePrescriptionRequest{
		PatientID: req.PatientId,
		DrugName:  req.DrugName,
		Dosage:    req.Dosage,
	}

	err = s.dispenseSvc.CreatePrescription(ctx, doctorID, input)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Return more useful data
	return &pharm.PrescriptionResponse{
		Status:  "received",
		Message: "Prescription created successfully",
	}, nil
}
