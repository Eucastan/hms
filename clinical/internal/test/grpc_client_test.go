package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eucastan/hms/clinical/internal/grpc/client"
	"github.com/Eucastan/hms/clinical/internal/handlers"
	"github.com/Eucastan/hms/clinical/internal/models"
	"github.com/Eucastan/hms/shared/pkg/auth"
	labpb "github.com/Eucastan/hms/shared/pkg/proto/lab"
	patient "github.com/Eucastan/hms/shared/pkg/proto/patient"
	pharm "github.com/Eucastan/hms/shared/pkg/proto/pharmacy"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	//"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// MOCK SERVERS

type mockPatientServer struct {
	patient.UnimplementedPatientServiceServer
}

func (s *mockPatientServer) GetPatient(ctx context.Context, req *patient.GetPatientRequest) (*patient.GetPatientResponse, error) {
	return &patient.GetPatientResponse{
		PatientInfo: &patient.Patient{
			HospitalNo:    "TEST-HNO-001",
			FirstName:     "Test",
			LastName:      "Patient",
			Gender:        "Male",
			Address:       "Test Address, Lagos",
			Phone:         "08012345678",
			NextOfKinName: "Test Kin",
		},
	}, nil
}

func (s *mockPatientServer) SearchPatients(ctx context.Context, req *patient.SearchPatientsRequest) (*patient.SearchPatientsResponse, error) {
	return &patient.SearchPatientsResponse{
		Patients: []*patient.Patient{
			{
				HospitalNo:    "TEST-HNO-001",
				FirstName:     "Test",
				LastName:      "Patient",
				Gender:        "Male",
				Address:       "Test Address, Lagos",
				Phone:         "08012345678",
				NextOfKinName: "Test Kin",
			},
		},
	}, nil
}

type mockLabServer struct {
	labpb.UnimplementedLabServiceServer
}

func (s *mockLabServer) CreateLabTestRequest(ctx context.Context, req *labpb.CreateLabRequest) (*labpb.LabResponse, error) {
	return &labpb.LabResponse{
		Status:  "success",
		Message: "Lab request created successfully",
	}, nil
}

func (s *mockLabServer) UpdateLabTestRequest(ctx context.Context, req *labpb.UpdateLabRequest) (*labpb.LabResponse, error) {
	return &labpb.LabResponse{
		Status:  "success",
		Message: "Lab request updated successfully",
	}, nil
}

type mockPharmacyServer struct {
	pharm.UnimplementedPharmacyServiceServer
}

func (s *mockPharmacyServer) CreatePrescription(ctx context.Context, req *pharm.PrescriptionRequest) (*pharm.PrescriptionResponse, error) {
	return &pharm.PrescriptionResponse{
		Status:  "success",
		Message: "Prescription created successfully",
	}, nil
}

// TEST SUITE

type GRPCClientHandlerTestSuite struct {
	suite.Suite
	lis     *bufconn.Listener
	grpcSrv *grpc.Server
	conn    *grpc.ClientConn
	router  *gin.Engine
	token   string
}

func (s *GRPCClientHandlerTestSuite) SetupSuite() {
	// Generate valid JWT
	doctorID := uuid.NewString()
	token, err := utils.GenerateToken(doctorID, "doctor", "test-secret-very-long-32-chars-min")
	s.Require().NoError(err)
	s.token = token

	// In-memory gRPC test server
	s.lis = bufconn.Listen(bufSize)
	s.grpcSrv = grpc.NewServer()

	// Register mocks with correct types
	labpb.RegisterLabServiceServer(s.grpcSrv, &mockLabServer{})
	pharm.RegisterPharmacyServiceServer(s.grpcSrv, &mockPharmacyServer{})
	patient.RegisterPatientServiceServer(s.grpcSrv, &mockPatientServer{})

	// Start server in background
	go func() {
		if err := s.grpcSrv.Serve(s.lis); err != nil && err != grpc.ErrServerStopped {
			panic(err)
		}
	}()

	// Dial in-memory connection
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return s.lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // ensure connection is ready
	)
	s.Require().NoError(err, "failed to create bufconn client")
	s.conn = conn

	// Create test client
	testClient := client.NewTestSendToAllClients(conn, conn, conn)
	grpcHandler := handlers.NewGRPCClientHandler(*testClient)

	// Setup Gin router (same as production)
	r := gin.Default()
	mw := auth.AuthMiddleware("test-secret-very-long-32-chars-min")
	c := r.Group("/api/v1", mw, auth.RequiredRole("doctor"))
	{
		c.POST("/prescription", auth.RequiredRole("doctor", "pharmacist"), grpcHandler.CreatePrescription)
		c.POST("/lab-request", auth.RequiredRole("doctor", "labtech"), grpcHandler.CreateLabRequest)
		c.PUT("/lab-request/:id", auth.RequiredRole("doctor", "labtech"), grpcHandler.UpdateLabRequest)
		c.GET("/patient/:id", auth.RequiredRole("doctor", "admin"), grpcHandler.GetPatient)
		c.GET("/patients/search", auth.RequiredRole("doctor", "admin"), grpcHandler.SearchPatient)
	}
	s.router = r
}

func (s *GRPCClientHandlerTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcSrv != nil {
		s.grpcSrv.Stop()
	}
	if s.lis != nil {
		s.lis.Close()
	}
}

func (s *GRPCClientHandlerTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var jsonBody []byte
	if body != nil {
		var err error
		jsonBody, err = json.Marshal(body)
		s.Require().NoError(err)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// TESTS

func (s *GRPCClientHandlerTestSuite) TestCreatePrescription() {
	body := models.CreatePrescriptionRequest{
		PatientID: uuid.NewString(),
		DoctorID:  "doctor-id-from-jwt",
		DrugName:  "Paracetamol",
		Dosage:    "500mg twice daily",
	}
	w := s.performRequest("POST", "/api/v1/prescription", body)

	s.Equal(http.StatusCreated, w.Code)
	s.Contains(w.Body.String(), "Prescription sent successfully")
}

func (s *GRPCClientHandlerTestSuite) TestCreateLabRequest() {
	body := models.LabRequest{
		PatientID: uuid.NewString(),
		RequestBy: "doctor-id-from-jwt", // will be set by handler
		TestType:  "Complete Blood Count",
		Priority:  "routine",
		Notes:     "Routine pre-op lab",
	}
	w := s.performRequest("POST", "/api/v1/lab-request", body)

	s.Equal(http.StatusCreated, w.Code)
	s.Contains(w.Body.String(), "Lab request sent successfully")
}

func (s *GRPCClientHandlerTestSuite) TestUpdateLabRequest() {
	labReqID := uuid.NewString()
	body := map[string]interface{}{
		"test_type": "Updated Blood Panel",
		"priority":  "urgent",
		"notes":     "Updated note from doctor",
	}
	w := s.performRequest("PUT", "/api/v1/lab-request/"+labReqID, body)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "Update successful")
}

func (s *GRPCClientHandlerTestSuite) TestGetPatient() {
	patientID := uuid.New()
	w := s.performRequest("GET", "/api/v1/patient/"+patientID.String(), nil)

	s.Equal(http.StatusOK, w.Code)

	var p models.Patient
	err := json.Unmarshal(w.Body.Bytes(), &p)
	s.Require().NoError(err)

	s.Equal("TEST-HNO-001", p.HospitalNo)
	s.Equal("Test", p.FirstName)
	s.Equal("Patient", p.LastName)
}

func (s *GRPCClientHandlerTestSuite) TestSearchPatients() {
	w := s.performRequest("GET", "/api/v1/patients/search?name=Test", nil)

	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	patients, ok := resp["patients"].([]interface{})
	s.Require().True(ok)
	s.Equal(1, len(patients))
	s.Equal(1, int(resp["count"].(float64)))
}

func TestGRPCClientHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCClientHandlerTestSuite))
}
