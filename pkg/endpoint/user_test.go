package endpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/PedPet/api-gateway/config"
	userSvcEndpoint "github.com/PedPet/user/pkg/endpoint"
	"github.com/bxcodec/faker/v3"
)

var username string = faker.Username()
var password string = faker.Password() + "1!"
var uSvc grpcSvcConnFunc
var expectedConfirmResponse string = `{"ok":true}`
var expectedConfirmResponseFalse string = `{"ok":false}`

func init() {
	start := time.Now()

	s, err := config.LoadSettings()
	if err != nil {
		log.Fatalf("Failed to load settings: %s", err)
	}

	svcFunc := userService(s.User)
	uSvc = svcFunc

	elapsed := time.Since(start)
	log.Printf("init took %s\n", elapsed)
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Log("username:", username)
	t.Log("password:", password)

	// Create request body / json
	udr := userSvcEndpoint.CreateUserRequest{
		Username: username,
		Email:    "scrott@gmail.com",
		Password: password,
	}

	data, err := json.Marshal(udr)
	if err != nil {
		t.Fatalf("Failed to encode create user request: %s", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "/user/register", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeRegisterEndpoint(ctx, uSvc))
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if respBody := rr.Body.String(); respBody != expectedConfirmResponse {
		t.Fatalf("handler returned unexpected body: got %v want %v", respBody, expectedConfirmResponse)
	}
}

func TestConfirmUser(t *testing.T) {
	t.Skip("Needs settings up with emailed user otp")
	ctx := context.Background()

	cur := userSvcEndpoint.ConfirmUserRequest{
		Username: "wAKLvTy",
		Code:     "274706",
	}
	data, err := json.Marshal(cur)
	if err != nil {
		t.Fatalf("Failed to encode confirm user request: %s", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "/user/confirm", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeConfirmUserEndpoint(ctx, uSvc))
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if respBody := rr.Body.String(); respBody != expectedConfirmResponse {
		t.Fatalf("handler returned unexpected body: got %v want %v", respBody, expectedConfirmResponse)
	}
}

func TestResendConfirmation(t *testing.T) {
	ctx := context.Background()

	rcr := userSvcEndpoint.ResendConfirmationRequest{
		Username: username,
	}
	data, err := json.Marshal(rcr)
	if err != nil {
		t.Fatalf("Failed to encode resend confirmation request: %s", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "/user/resend-confirmation", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")

	t.Logf("data: %s\n", data)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeResendConfirmationEndpoint(ctx, uSvc))
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if respBody := rr.Body.String(); respBody != expectedConfirmResponse {
		t.Fatalf("handler, returned unexpected body: got %v want %v", respBody, expectedConfirmResponse)
	}
}

func TestUsernameTaken(t *testing.T) {
	ctx := context.Background()

	// Create request
	req, err := http.NewRequest("GET", "/user/check-username-taken?username="+username, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("username: %s\n", username)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeCheckUsernameTakenEndpoint(ctx, uSvc))
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if respBody := rr.Body.String(); respBody != expectedConfirmResponse {
		t.Fatalf("handler, returned unexpected body: got %v want %v", respBody, expectedConfirmResponse)
	}
}

func TestLogin(t *testing.T) {
	// t.Skip("Needs settings up with emailed user otp")

	ctx := context.Background()
	expectedLoginResponse := "{\"jwt\":\"([\\d\\w.-]+)\"}"

	lr := userSvcEndpoint.LoginRequest{
		Username: "wAKLvTy",
		Password: "vQlTfoYVtGWEHyUjsbuxeJIFaETzEOnMvUxGjzMmXTQmyMmNWv1!",
	}
	data, err := json.Marshal(lr)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("username: %s\n", username)

	// Create request
	req, err := http.NewRequest("POST", "/user/login", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeLoginEndpoint(ctx, uSvc))
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if respBody := rr.Body.String(); !regexp.MustCompile(expectedLoginResponse).MatchString(respBody) {
		t.Errorf("handler, returned unexpected body: got %v want %s", respBody, expectedLoginResponse)
	}
}
