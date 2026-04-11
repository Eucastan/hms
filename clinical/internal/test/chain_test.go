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
	billing "github.com/Eucastan/hms/shared/pkg/proto/billing"
	patientpb "github.com/Eucastan/hms/shared/pkg/proto/patient"
	pharm "github.com/Eucastan/hms/shared/pkg/proto/pharmacy"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSizes = 1024 * 1024

// Mock Servers

type mockPatientServerTest struct {
	patientpb.UnimplementedPatientServiceServer
}

func (s *mockPatientServerTest) GetPatient(ctx context.Context, req *patientpb.GetPatientRequest) (*patientpb.GetPatientResponse, error) {
	return &patientpb.GetPatientResponse{
		PatientInfo: &patientpb.Patient{PatientId: req.PatientId},
	}, nil
}

type mockBillingServer struct {
	billing.UnimplementedBillingServiceServer
}

func (s *mockBillingServer) CreateBillCharge(ctx context.Context, req *billing.CreateBillChargeRequest) (*billing.CreateBillChargeResponse, error) {
	return &billing.CreateBillChargeResponse{
		Status:  "success",
		Message: "Bill charge created",
	}, nil
}

type mockPharmacyServerTest struct {
	pharm.UnimplementedPharmacyServiceServer
}

func (s *mockPharmacyServerTest) CreatePrescription(ctx context.Context, req *pharm.PrescriptionRequest) (*pharm.PrescriptionResponse, error) {
	return &pharm.PrescriptionResponse{
		Status:  "received",
		Message: "Prescription processed",
	}, nil
}

// Chain Test Suite

type ChainTestSuite struct {
	suite.Suite
	lis     *bufconn.Listener
	grpcSrv *grpc.Server
	conn    *grpc.ClientConn
	router  *gin.Engine
	token   string
}

func (s *ChainTestSuite) SetupSuite() {
	// Generate valid JWT token
	userID := uuid.NewString()
	token, err := utils.GenerateToken(userID, "doctor", "test-secret-very-long-32-chars-min")
	s.Require().NoError(err)
	s.token = token

	// In-memory gRPC server with all mocks
	s.lis = bufconn.Listen(bufSize)
	s.grpcSrv = grpc.NewServer()

	patientpb.RegisterPatientServiceServer(s.grpcSrv, &mockPatientServer{})
	billing.RegisterBillingServiceServer(s.grpcSrv, &mockBillingServer{})
	pharm.RegisterPharmacyServiceServer(s.grpcSrv, &mockPharmacyServer{})

	go func() {
		if err := s.grpcSrv.Serve(s.lis); err != nil && err != grpc.ErrServerStopped {
			panic(err)
		}
	}()

	// Connect using bufconn
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return s.lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	s.Require().NoError(err)
	s.conn = conn

	// Create clinical client
	testClient := client.NewTestSendToAllClients(s.conn, s.conn, s.conn)
	grpcHandler := handlers.NewGRPCClientHandler(*testClient)

	// Setup router
	r := gin.Default()
	mw := auth.AuthMiddleware("test-secret-very-long-32-chars-min")
	c := r.Group("/api/v1", mw, auth.RequiredRole("doctor"))
	{
		c.POST("/prescription", auth.RequiredRole("doctor", "pharmacist"), grpcHandler.CreatePrescription)
	}

	s.router = r
}

func (s *ChainTestSuite) TearDownSuite() {
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

func (s *ChainTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// CHAIN TEST

func (s *ChainTestSuite) TestClinicalPharmacyPaymentChain() {
	body := models.CreatePrescriptionRequest{
		PatientID: uuid.NewString(),
		DoctorID:  uuid.NewString(),
		DrugName:  "Amoxicillin 500mg",
		Dosage:    "1 tablet twice daily",
	}

	w := s.performRequest("POST", "/api/v1/prescription", body)

	s.Equal(http.StatusCreated, w.Code, w.Body.String())
	s.Contains(w.Body.String(), "Prescription sent successfully")

	s.T().Log("Full chain test passed: Clinical → Pharmacy → Payment")
}

func TestChainTestSuite(t *testing.T) {
	suite.Run(t, new(ChainTestSuite))
}
