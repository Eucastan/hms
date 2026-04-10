package test

import (
	"context"

	patient "github.com/Eucastan/hms/shared/pkg/proto/patient"
)

type FakePatientServer struct {
	patient.UnimplementedPatientServiceServer
}

func (f *FakePatientServer) GetPatient(
	ctx context.Context,
	req *patient.GetPatientRequest,
) (*patient.GetPatientResponse, error) {

	return &patient.GetPatientResponse{
		PatientInfo: &patient.Patient{
			PatientId: req.PatientId,
		},
	}, nil
}
