package server

import (
	"context"
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eucastan/hms/patient/internal/services"
	patient "github.com/Eucastan/hms/shared/pkg/proto/patient"

	"github.com/google/uuid"
)

type PatientServer struct {
	patient.UnimplementedPatientServiceServer
	svc services.PatientService
}

func NewPatientServer(svc services.PatientService) *PatientServer {
	return &PatientServer{svc: svc}
}

func (s *PatientServer) GetPatient(ctx context.Context, req *patient.GetPatientRequest) (*patient.GetPatientResponse, error) {
	patientID, err := uuid.Parse(req.PatientId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid patient id format")
	}

	pts, err := s.svc.FindByID(ctx, patientID)
	if err != nil {
		if errors.Is(err, errors.New("patient not found")) {
			return nil, status.Error(codes.NotFound, "patient not found")
		}
		// Log internal error, return generic to client
		log.Printf("GetPatient failed for %s: %v", req.PatientId, err)
		return nil, status.Error(codes.Internal, "failed to retrieve patient record")
	}

	return &patient.GetPatientResponse{
		PatientInfo: &patient.Patient{
			PatientId:     pts.ID.String(),
			HospitalNo:    pts.HospitalNo,
			FirstName:     pts.FirstName,
			LastName:      pts.LastName,
			Gender:        pts.Gender,
			Address:       pts.Address,
			Phone:         pts.Phone,
			NextOfKinName: pts.NextOfKinName,
		},
	}, nil
}

func (s *PatientServer) SearchPatients(ctx context.Context, req *patient.SearchPatientsRequest) (*patient.SearchPatientsResponse, error) {

	if req.Name == "" && req.HospitalNo == "" {
		return nil, status.Error(codes.InvalidArgument, "at least one of name or hospital_no is required")
	}

	patients, err := s.svc.SearchPatient(ctx, req.Name, req.HospitalNo)
	if err != nil {
		log.Printf("SearchPatients failed: %v", err)
		return nil, status.Error(codes.Internal, "Search failed")
	}

	var PatientInfo []*patient.Patient
	for _, p := range patients {
		PatientInfo = append(PatientInfo, &patient.Patient{
			PatientId:     p.ID.String(),
			HospitalNo:    p.HospitalNo,
			FirstName:     p.FirstName,
			LastName:      p.LastName,
			Gender:        p.Gender,
			Address:       p.Address,
			Phone:         p.Phone,
			NextOfKinName: p.NextOfKinName,
		})
	}

	return &patient.SearchPatientsResponse{
		Patients: PatientInfo,
	}, nil
}
