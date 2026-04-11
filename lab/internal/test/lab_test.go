package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eucastan/hms/lab/internal/models"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

var userID string

type LabHandlerTestSuite struct {
	suite.Suite
	server *TestServer
	token  string
}

func (s *LabHandlerTestSuite) SetupSuite() {
	db := SetupTestDB(s.T())
	s.server = NewLabTestServer(s.T(), db)

	userID = uuid.NewString()
	token, err := utils.GenerateToken(userID, "lab-tech", "test-secret-very-long-32-chars-min")
	s.Require().NoError(err)
	s.token = token
}

func (s *LabHandlerTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *LabHandlerTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *LabHandlerTestSuite) TestLabRequest() {

	patientID := uuid.NewString()

	labRequest := &models.LabRequestCreate{
		PatientID: patientID,
		TestType:  "malaria",
		Priority:  "routine",
		Notes:     "This is a note from lab",
	}

	w := s.performRequest("POST", "/api/v1/lab-request", labRequest)
	s.T().Log(w.Body.String())
	s.Require().Equal(http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	labReqData := resp["labreq"].(map[string]interface{})
	idVal, ok := labReqData["id"]
	s.Require().True(ok, "id missing in response")

	labReqID := idVal.(string)

	// Get Diagnosis
	w = s.performRequest("GET", "/api/v1/lab-request/"+patientID, nil)
	s.Equal(http.StatusOK, w.Code)
	s.T().Log(w.Body.String())

	// Delete Diagnosis
	w = s.performRequest("DELETE", "/api/v1/lab-request/"+labReqID, nil)
	s.Equal(http.StatusOK, w.Code)
	s.T().Log(w.Body.String())
}

func (s *LabHandlerTestSuite) TestLabResult() {

	patientID := uuid.NewString()

	labRequest := &models.LabResultCreate{
		PatientID:      patientID,
		TestType:       "malaria",
		ResultValue:    "malaria parasite",
		Unit:           "",
		ReferenceRange: "",
		Comments:       "patient heart rate is not stable",
		Verified:       true,
	}

	w := s.performRequest("POST", "/api/v1/lab-result", labRequest)
	s.T().Log(w.Body.String())
	s.Require().Equal(http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	labReqData := resp["result"].(map[string]interface{})
	idVal, ok := labReqData["id"]
	s.Require().True(ok, "id missing in response")

	labResultID := idVal.(string)

	// Get Diagnosis
	w = s.performRequest("GET", "/api/v1/lab-result/"+patientID, nil)
	s.Equal(http.StatusOK, w.Code)
	s.T().Log(w.Body.String())

	// Delete Diagnosis
	w = s.performRequest("DELETE", "/api/v1/lab-result/"+labResultID, nil)
	s.Equal(http.StatusOK, w.Code)
	s.T().Log(w.Body.String())
}

func TestLabHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LabHandlerTestSuite))
}
