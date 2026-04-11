package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eucastan/hms/clinical/internal/dto"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ClinicalHandlerTestSuite struct {
	suite.Suite
	server *TestServer
	token  string
}

var doctorID string

func (s *ClinicalHandlerTestSuite) SetupSuite() {
	db := SetupTestDB(s.T())
	s.server = NewClinicalTestServer(s.T(), db)

	doctorID = uuid.NewString()
	token, err := utils.GenerateToken(doctorID, "doctor", "test-secret-very-long-32-chars-min")
	s.Require().NoError(err)
	s.token = token
}

func (s *ClinicalHandlerTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *ClinicalHandlerTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *ClinicalHandlerTestSuite) TestClinical() {
	patientID := uuid.New()

	diagnosis := &dto.DiagnosisCreateRequest{
		PatientID:   patientID,
		AdmissionID: uuid.New(),
		Code:        "54-007-10CD",
		Description: "This is description",
	}

	w := s.performRequest("POST", "/api/v1/diagnosis", diagnosis)
	s.T().Log(w.Body.String())
	s.Require().Equal(http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	idVal, ok := resp["id"]
	s.Require().True(ok, "id missing in response")

	diagID, ok := idVal.(string)
	s.Require().True(ok, "id is not a string")
	s.T().Log("Extracted diagID:", diagID)

	// Get Diagnosis
	w = s.performRequest("GET", "/api/v1/diagnosis/"+diagID, nil)
	s.T().Log("GET Response:", w.Body.String())
	s.Equal(http.StatusOK, w.Code)

	// Delete Diagnosis
	w = s.performRequest("DELETE", "/api/v1/diagnosis/"+diagID, nil)
	s.T().Log("DELETE Response:", w.Body.String())
	s.Equal(http.StatusOK, w.Code)
}

// TestUpdateDiagnosis tests the full update flow
func (s *ClinicalHandlerTestSuite) TestUpdateDiagnosis() {
	// Create first
	patientID := uuid.New()
	createReq := &dto.DiagnosisCreateRequest{
		PatientID:   patientID,
		AdmissionID: uuid.New(),
		Code:        "54-007-10CD",
		Description: "Initial description for patient",
	}

	w := s.performRequest("POST", "/api/v1/diagnosis", createReq)
	s.Require().Equal(http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	diagID := createResp["id"].(string)

	s.T().Log("Created diagnosis ID:", diagID)

	// Update
	updateReq := map[string]string{
		"code":        "54-008-20EF",
		"description": "Updated description - patient condition improved significantly",
	}

	w = s.performRequest("PUT", "/api/v1/diagnosis/"+diagID, updateReq)
	s.T().Log("UPDATE Response:", w.Body.String())
	s.Require().Equal(http.StatusOK, w.Code)

	// Verify update
	w = s.performRequest("GET", "/api/v1/diagnosis/"+diagID, nil)
	s.Require().Equal(http.StatusOK, w.Code)

	var getResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &getResp)
	diagnosis := getResp["diagnosis"].(map[string]interface{})

	s.Require().Equal("54-008-20EF", diagnosis["Code"])

	s.T().Log("Update test passed")
}

func TestClinicalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ClinicalHandlerTestSuite))
}
