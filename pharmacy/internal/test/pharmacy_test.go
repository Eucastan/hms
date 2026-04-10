package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eucastan/hms/pharmacy/internal/models"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type PharmacyHandlerTestSuite struct {
	suite.Suite
	server *TestServer
	token  string
}

func (s *PharmacyHandlerTestSuite) SetupSuite() {
	db := SetupTestDB(s.T())
	s.server = NewPharmacyTestServer(s.T(), db)

	userID := uuid.NewString()

	token, err := utils.GenerateToken(userID, "pharmacist", "test-secret-very-long-32-chars-min")
	s.Require().NoError(err)
	s.token = token
}

func (s *PharmacyHandlerTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *PharmacyHandlerTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *PharmacyHandlerTestSuite) TestPharmacyDrug() {
	var drugID string

	createDrugBody := &models.DrugCreateRequest{
		Name:         "Amoxicilin",
		GenericName:  "IDoNotKnowWhatTheHellThatIs",
		Form:         "tablet",
		Strength:     "400mg",
		PackSize:     40,
		Stock:        55,
		PricePerUnit: 2500.00,
		MinStock:     10,
	}

	w := s.performRequest("POST", "/api/v1/drug", createDrugBody)

	s.Equal(http.StatusCreated, w.Code)
	if w.Code != http.StatusCreated {
		s.T().Log(w.Body.String())
		s.FailNow("Create failed")
	}

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	drugData := resp["drug"].(map[string]interface{})
	idVal, ok := drugData["id"]
	s.Require().True(ok, "id missing in response")

	drugID = idVal.(string)

	// Get Drug
	w = s.performRequest("GET", "/api/v1/drug/"+drugID, nil)
	s.Equal(http.StatusOK, w.Code)

	// Delete Drug
	w = s.performRequest("DELETE", "/api/v1/drug/"+drugID, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "Drug deleted")

}

func (s *PharmacyHandlerTestSuite) createTestDrug() string {
	createDrugBody := &models.DrugCreateRequest{
		Name:         "Ciprofloxacin-" + uuid.NewString(),
		GenericName:  "TestGeneric",
		Form:         "tablet",
		Strength:     "500mg",
		PackSize:     40,
		Stock:        155,
		PricePerUnit: 2500.00,
		MinStock:     10,
	}
	w := s.performRequest("POST", "/api/v1/drug", createDrugBody)
	s.Require().Equal(http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	drugData := resp["drug"].(map[string]interface{})
	idVal := drugData["id"].(string)
	return idVal
}

func (s *PharmacyHandlerTestSuite) TestPharmacyDispense() {
	drugID := s.createTestDrug()
	patientID := uuid.NewString()

	dispense := &models.CreateDispenseRequest{
		PatientID:      patientID,
		PrescriptionID: "",
		DrugID:         drugID,
		Quantity:       2,
		Notes:          "Taken 2 times daily M&N",
	}

	w := s.performRequest("POST", "/api/v1/dispense", dispense)
	s.Equal(http.StatusCreated, w.Code)
	s.Require().Equal(http.StatusCreated, w.Code, w.Body.String())

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	dispenseData := resp["dispense"].(map[string]interface{})
	idVal, ok := dispenseData["id"]
	s.Require().True(ok, "id missing in response")

	dispenseID := idVal.(string)

	// Get Dispense
	w = s.performRequest("GET", "/api/v1/dispense/"+dispenseID, nil)
	s.Equal(http.StatusOK, w.Code)

}

func (s *PharmacyHandlerTestSuite) TestPharmacyMockDispense() {
	drugID := s.createTestDrug()
	patientID := uuid.NewString()

	dispense := &models.CreateDispenseRequest{
		PatientID:      patientID,
		PrescriptionID: "",
		DrugID:         drugID,
		Quantity:       1,
		Notes:          "Taken 2 times daily M&N",
	}

	w := s.performRequest("POST", "/api/v1/mock/dispense", dispense)
	s.Equal(http.StatusCreated, w.Code)
	s.Require().Equal(http.StatusCreated, w.Code, w.Body.String())

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	dispenseData := resp["dispense"].(map[string]interface{})
	idVal, ok := dispenseData["id"]
	s.Require().True(ok, "id missing in response")

	dispenseID := idVal.(string)

	// Get Dispense
	w = s.performRequest("GET", "/api/v1/mock/dispense/"+dispenseID, nil)
	s.Equal(http.StatusOK, w.Code)

}

func TestPharmacyHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PharmacyHandlerTestSuite))
}
