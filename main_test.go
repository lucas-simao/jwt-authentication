package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

func TestSignUp(t *testing.T) {
	tests := map[string]struct {
		auth             Authentication
		expectedCode     int
		expectedResponde Response
	}{
		"1 - should return 201": {
			auth: Authentication{
				Username: "lucas",
				Password: "12345678",
			},
			expectedCode:     http.StatusCreated,
			expectedResponde: Response{"successfully"},
		},
		"2 - should return 400 - without username": {
			auth: Authentication{
				Username: "",
				Password: "12345678",
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponde: Response{"inform username and password"},
		},
		"3 - should return 400 - without password": {
			auth: Authentication{
				Username: "lucas",
				Password: "",
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponde: Response{"inform username and password"},
		},
		"4 - should return 400 - without auth": {
			auth: Authentication{
				Username: "",
				Password: "",
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponde: Response{"inform username and password"},
		},
		"5 - should return 400 - user already exist": {
			auth: Authentication{
				Username: "lucas",
				Password: "12345678",
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponde: Response{"sorry, that username already exist"},
		},
	}

	keys := make([]string, 0, len(tests))

	for k := range tests {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		t.Run(k, func(t *testing.T) {
			test := tests[k]
			var buffer bytes.Buffer
			err := json.NewEncoder(&buffer).Encode(test.auth)
			if err != nil {
				t.Errorf("error on encoder payload: %v", err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/singup", &buffer)
			w := httptest.NewRecorder()
			f := SignUp()
			f.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != test.expectedCode {
				t.Errorf("expect: %d but got: %d error: %v", test.expectedCode, resp.StatusCode, err)
				return
			}

			var response Response

			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				t.Errorf("error on decoder payload: %v", err)
				return
			}

			if response.Message != test.expectedResponde.Message {
				t.Errorf("expect: %s but got %s", test.expectedResponde.Message, response.Message)
			}
		})
	}
}

func TestSingIn(t *testing.T) {
	username, password := "test", "test"
	DB[username] = password

	tests := map[string]struct {
		auth             Authentication
		expectedCode     int
		expectedResponde Response
	}{
		"should return 200": {
			auth:         Authentication{Username: username, Password: password},
			expectedCode: http.StatusOK,
		},
		"should return 401": {
			auth:             Authentication{Username: "test1", Password: password},
			expectedCode:     http.StatusUnauthorized,
			expectedResponde: Response{"wrong username or password"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := bytes.Buffer{}
			err := json.NewEncoder(&buffer).Encode(test.auth)
			if err != nil {
				t.Errorf("error on encoder payload: %v", err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/singin", &buffer)
			w := httptest.NewRecorder()
			SingIn().ServeHTTP(w, req)
			resp := w.Result()

			defer resp.Body.Close()

			if resp.StatusCode != test.expectedCode {
				t.Errorf("expect: %d but got: %d error: %v", test.expectedCode, resp.StatusCode, err)
				return
			}

			var response Response

			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				t.Errorf("error on decoder payload: %v", err)
				return
			}

			if resp.StatusCode != 200 && response.Message != test.expectedResponde.Message {
				t.Errorf("expect: %s but got %s", test.expectedResponde.Message, response.Message)
				return
			}
		})
	}
}

func TestWelcome(t *testing.T) {
	username, password := "welcome", "welcome"
	DB[username] = password

	req := httptest.NewRequest(http.MethodPost, "/singin", strings.NewReader(`{"username": "welcome", "password":"welcome"}`))
	w := httptest.NewRecorder()
	SingIn().ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("error to create sing-in user expect: %d but got: %d", http.StatusOK, w.Result().StatusCode)
		return
	}

	response := map[string]string{}

	err := json.NewDecoder(w.Result().Body).Decode(&response)
	if err != nil {
		t.Errorf("error on decoder payload: %v", err)
		return
	}

	tests := map[string]struct {
		token            string
		expectedCode     int
		expectedResponde Response
	}{
		"should return 200": {
			token:            response["token"],
			expectedCode:     http.StatusOK,
			expectedResponde: Response{"Welcome!"},
		},
		"should return 401 - empty token": {
			token:            "",
			expectedCode:     http.StatusUnauthorized,
			expectedResponde: Response{"token is malformed: token contains an invalid number of segments"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req = httptest.NewRequest(http.MethodGet, "/welcome", nil)
			req.Header.Set("Authorization", "Bearer "+test.token)
			w = httptest.NewRecorder()
			Welcome().ServeHTTP(w, req)

			var respWelcome Response

			err = json.NewDecoder(w.Result().Body).Decode(&respWelcome)
			if err != nil {
				t.Errorf("error on decoder payload: %v", err)
				return
			}

			if respWelcome.Message != test.expectedResponde.Message {
				t.Errorf("error on get response expect: %s but got: %s", test.expectedResponde.Message, respWelcome.Message)
			}
		})
	}
}
