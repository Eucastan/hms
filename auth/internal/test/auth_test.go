package test

import (
	"bytes"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/Eucastan/hms/auth/internal/models"
	"github.com/Eucastan/hms/shared/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	server *TestServer
	token  string
}

func (s *AuthHandlerTestSuite) SetupSuite() {
	db := SetupTestDB(s.T())
	s.server = NewAuthServer(s.T(), db)

	userID := uuid.NewString()
	token, err := utils.GenerateToken(userID, "admin", "testsecret")
	s.Require().NoError(err)
	s.token = token
}

func (s *AuthHandlerTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AuthHandlerTestSuite) TestAuth() {

	registerBody := models.StaffCreateRequest{
		Email:     "test@example.com",
		Password:  "VeryStrongPassword!123#Secure",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "admin",
	}

	w := s.performRequest("POST", "/register", registerBody)

	if w.Code != http.StatusCreated {
		s.T().Fatalf("register failed: %s", w.Body.String())
	}

	loginBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "VeryStrongPassword!123#Secure",
	}

	w = s.performRequest("POST", "/login", loginBody)
	if w.Code != http.StatusOK {
		s.T().Fatalf("login failed: %s", w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	token := resp["token"]
	if token == nil {
		s.T().Fatalf("expected JWT token")
	}
}

func (s *AuthHandlerTestSuite) TearDownSuite() {
	s.server.Close()
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
