package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Eucastan/hms/patient/internal/models"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type PatientHandlerTestSuite struct {
	suite.Suite
	server *TestServer
	token  string
}

func (s *PatientHandlerTestSuite) SetupSuite() {
	db := SetupTestDB(s.T())
	s.server = NewPatientTestServer(s.T(), db)

	token, err := utils.GenerateToken("test-user-id", "admin", "test-secret-very-long-32-chars-min")
	s.Require().NoError(err)
	s.token = token
}

func (s *PatientHandlerTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *PatientHandlerTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *PatientHandlerTestSuite) TestCreateAndGetPatient() {

	hospitalNo := strings.ToUpper("HMS-TEST-" + uuid.NewString())

	patient := &models.PatientCreate{
		HospitalNo:    hospitalNo,
		FirstName:     "Test",
		LastName:      "Patient",
		DateOfBirth:   "1995-03-20",
		Age:           30,
		Gender:        "M",
		Address:       "Lagos",
		Phone:         "08030000000",
		NextOfKinName: "Test Kin",
		InitialWard:   "GeneralWard",
		InitialReason: "accident",
	}

	w := s.performRequest("POST", "/api/v1/patient", patient)

	s.Equal(http.StatusCreated, w.Code)
	if w.Code != http.StatusCreated {
		s.T().Log(w.Body.String())
		s.FailNow("Create failed")
	}

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	patientData := resp["patient"].(map[string]interface{})
	idVal, ok := patientData["id"]
	s.Require().True(ok, "id missing in response")

	patientID := idVal.(string)

	w = s.performRequest("GET", "/api/v1/patient/"+patientID, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *PatientHandlerTestSuite) TestUpdatePatient() {

	hospitalNo := strings.ToUpper("HMS-TEST-" + uuid.NewString())

	patient := &models.PatientCreate{
		HospitalNo:    hospitalNo,
		FirstName:     "Test",
		LastName:      "Patient",
		DateOfBirth:   "1995-03-20",
		Age:           30,
		Gender:        "M",
		Address:       "Lagos",
		Phone:         "08030000000",
		NextOfKinName: "Test Kin",
		InitialWard:   "GeneralWard",
		InitialReason: "accident",
	}

	w := s.performRequest("POST", "/api/v1/patient", patient)

	s.Equal(http.StatusCreated, w.Code)
	if w.Code != http.StatusCreated {
		s.T().Log(w.Body.String())
		s.FailNow("Create failed")
	}

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	patientData := resp["patient"].(map[string]interface{})
	idVal, ok := patientData["id"]
	if !ok || idVal == nil {
		s.T().Logf("response: %+v", resp)
		s.FailNow("id not found in response")
	}

	patientID := idVal.(string)

	updatePatient := map[string]interface{}{
		"first_name": "Updated",
		"age":        35,
	}

	w = s.performRequest("PUT", "/api/v1/patient/"+patientID, updatePatient)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "Update successful")

	w = s.performRequest("DELETE", "/api/v1/patient/"+patientID, nil)
	s.Equal(http.StatusOK, w.Code)
}

func TestPatientHandlerSuite(t *testing.T) {
	suite.Run(t, new(PatientHandlerTestSuite))
}
