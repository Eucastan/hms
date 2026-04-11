package client

import (
	"context"
	"github.com/Eucastan/hms/clinical/internal/models"
	labpb "github.com/Eucastan/hms/shared/pkg/proto/lab"
	patient "github.com/Eucastan/hms/shared/pkg/proto/patient"
	pharm "github.com/Eucastan/hms/shared/pkg/proto/pharmacy"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"time"
)

type SendToAllClients struct {
	labClient     labpb.LabServiceClient
	pharmClient   pharm.PharmacyServiceClient
	patientClient patient.PatientServiceClient
	retryConfig   string
}

func dial(Addr, retryConfig string) *grpc.ClientConn {
	Opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if retryConfig != "" {
		Opts = append(Opts, grpc.WithDefaultServiceConfig(retryConfig))
	}

	Conn, err := grpc.NewClient(Addr, Opts...)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", Addr, err)
	}

	return Conn
}

func NewSendToAllClient(patientAddr, labAddr, pharmAddr string) *SendToAllClients {
	retryCfgBytes, err := os.ReadFile("retry_policy.json")
	retryConfig := ""

	if err != nil {
		log.Printf("Warning: could not load retry-policy: %v", err)
	} else {
		retryConfig = string(retryCfgBytes)
	}

	return &SendToAllClients{
		labClient:     labpb.NewLabServiceClient(dial(labAddr, retryConfig)),
		pharmClient:   pharm.NewPharmacyServiceClient(dial(pharmAddr, retryConfig)),
		patientClient: patient.NewPatientServiceClient(dial(patientAddr, retryConfig)),
	}
}

func NewTestSendToAllClients(patientConn, labConn, pharmConn *grpc.ClientConn) *SendToAllClients {
	return &SendToAllClients{
		labClient:     labpb.NewLabServiceClient(labConn),
		pharmClient:   pharm.NewPharmacyServiceClient(pharmConn),
		patientClient: patient.NewPatientServiceClient(patientConn),
		retryConfig:   "", // not used in tests
	}
}

func (c *SendToAllClients) GetPatient(ctx context.Context, patientID uuid.UUID) (*models.Patient, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req := &patient.GetPatientRequest{
		PatientId: patientID.String(),
	}

	resp, err := c.patientClient.GetPatient(ctx, req)
	if err != nil {
		return nil, utils.ToPublicError(err)
	}

	if resp.PatientInfo == nil {
		return nil, utils.ErrNotFound
	}

	protoPatient := &models.Patient{
		HospitalNo:    resp.PatientInfo.HospitalNo,
		FirstName:     resp.PatientInfo.FirstName,
		LastName:      resp.PatientInfo.LastName,
		Gender:        resp.PatientInfo.Gender,
		Address:       resp.PatientInfo.Address,
		Phone:         resp.PatientInfo.Phone,
		NextOfKinName: resp.PatientInfo.NextOfKinName,
	}

	return protoPatient, nil
}

func (c *SendToAllClients) SearchPatients(ctx context.Context, name, hospitalNo string) ([]*models.Patient, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req := &patient.SearchPatientsRequest{
		Name:       name,
		HospitalNo: hospitalNo,
	}

	resp, err := c.patientClient.SearchPatients(ctx, req)
	if err != nil {
		return nil, utils.ToPublicError(err)
	}

	var patients []*models.Patient
	for _, p := range resp.Patients {
		patients = append(patients, &models.Patient{
			HospitalNo:    p.HospitalNo,
			FirstName:     p.FirstName,
			LastName:      p.LastName,
			Gender:        p.Gender,
			Address:       p.Address,
			Phone:         p.Phone,
			NextOfKinName: p.NextOfKinName,
		})
	}

	return patients, nil
}

func (c *SendToAllClients) SendPrescriptionToPharmacy(ctx context.Context, req *models.CreatePrescriptionRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	prescription := &pharm.PrescriptionRequest{
		PatientId: req.PatientID,
		DoctorId:  req.DoctorID,
		DrugName:  req.DrugName,
		Dosage:    req.Dosage,
	}

	_, err := c.pharmClient.CreatePrescription(ctx, prescription)
	return utils.ToPublicError(err)
}

func (c *SendToAllClients) SendLabTestRequest(ctx context.Context, req *models.LabRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	labReq := &labpb.CreateLabRequest{
		PatientId: req.PatientID,
		RequestBy: req.RequestBy,
		TestType:  req.TestType,
		Priority:  req.Priority,
		Status:    "requested",
		Notes:     req.Notes,
	}

	_, err := c.labClient.CreateLabTestRequest(ctx, labReq)
	return utils.ToPublicError(err)
}

func (c *SendToAllClients) UpdateLabRequest(ctx context.Context, id, requestBy string, req *models.LabRequestUpdate) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	labReq := &labpb.UpdateLabRequest{
		Id:        id,
		RequestBy: requestBy,
		Status:    "requested",
	}

	if req.TestType != nil {
		labReq.TestType = *req.TestType
	}

	if req.Priority != nil {
		labReq.Priority = *req.Priority
	}

	if req.Notes != nil {
		labReq.Notes = *req.Notes
	}

	_, err := c.labClient.UpdateLabTestRequest(ctx, labReq)
	return utils.ToPublicError(err)
}
