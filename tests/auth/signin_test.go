package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tibz-Dankan/keep-active/tests/setup"
)

func TestInvalidSignInCredentials(t *testing.T) {
	setup.ClearAllTables()

	var label string
	var payload []byte
	var req *http.Request
	var response *httptest.ResponseRecorder

	label = "Expects 400 with missing/invalid email"
	payload = []byte(`{"email":"","password":"password"}`)
	req, _ = http.NewRequest("POST", "/api/v1/auth/signin", bytes.NewBuffer(payload))
	response = setup.ExecuteRequest(req)
	setup.CheckResponseCode(t, label, http.StatusBadRequest, response.Code)

	label = "Expects 400 with missing/invalid password"
	payload = []byte(`{"email":"user@gmail.com","password":""}`)
	req, _ = http.NewRequest("POST", "/api/v1/auth/signin", bytes.NewBuffer(payload))
	response = setup.ExecuteRequest(req)
	setup.CheckResponseCode(t, label, http.StatusBadRequest, response.Code)
}

func TestSuccessfulSignIn(t *testing.T) {
	setup.ClearAllTables()

	var label string
	var payload []byte
	var req *http.Request
	var response *httptest.ResponseRecorder
	var body map[string]interface{}

	label = "Expects 200 on successful signup"
	payload = []byte(`{"name":"username","email":"user@gmail.com","password":"password"}`)
	req, _ = http.NewRequest("POST", "/api/v1/auth/signup", bytes.NewBuffer(payload))
	_ = setup.ExecuteRequest(req)

	payload = []byte(`{"email":"user@gmail.com","password":"password"}`)
	req, _ = http.NewRequest("POST", "/api/v1/auth/signin", bytes.NewBuffer(payload))
	response = setup.ExecuteRequest(req)
	setup.CheckResponseCode(t, label, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &body)

	label = "Expects accessToken on successful signin"
	token, found := body["accessToken"]
	if !found {
		fmt.Printf("=== FAIL: %s\n", label)
		t.Errorf("Expects accessToken. Got %v\n", token)
	}

	accessToken, ok := token.(string)
	if !ok {
		fmt.Printf("=== FAIL: %s\n", label)
		t.Errorf("Expects accessToken. Got %v\n", accessToken)
	}
	if accessToken == "" {
		fmt.Printf("=== FAIL: %s\n", label)
		t.Errorf("Expects accessToken. Got an empty string")
	} else {
		fmt.Printf("=== PASS: %s\n", label)
	}
}