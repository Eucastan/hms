package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eucastan/hms/payment/internal/models"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type PaymentHandlerTestSuite struct {
	suite.Suite
	server *TestServer
	token  string
}

func (s *PaymentHandlerTestSuite) SetupSuite() {
	db := SetupTestDB(s.T())
	s.server = NewPaymentTestServer(s.T(), db)

	userID := uuid.NewString()
	token, err := utils.GenerateToken(userID, "accountant", "test-secret-very-long-32-chars-min")
	s.Require().NoError(err)
	s.token = token
}

func (s *PaymentHandlerTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *PaymentHandlerTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.server.Router.ServeHTTP(w, req)

	return w
}

var refType string
var patientID string
var sourceRefID string

func (s *PaymentHandlerTestSuite) TestPayment() {

	patientID = uuid.NewString()
	sourceRefID = uuid.NewString()
	refType = "drug"

	billCharge := &models.BillChargeRequest{
		PatientID:     patientID,
		SourceRefID:   sourceRefID,
		ReferenceType: refType,
		Description:   "Payment for " + refType,
		Quantity:      4,
		UnitPrice:     5700.00,
	}

	w := s.performRequest("POST", "/api/v1/bill", billCharge)
	s.Equal(http.StatusCreated, w.Code)
	s.Require().Equal(http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	invoiceID, ok := resp["invoice_id"].(string)
	s.Require().True(ok, "invoice_id missing in response")

	// Get Invoice
	w = s.performRequest("GET", "/api/v1/invoice/"+invoiceID, nil)
	s.Equal(http.StatusOK, w.Code)
}

func TestPharmacyHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentHandlerTestSuite))
}
