package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var DB = map[string]string{}

var JWT_SECRET = []byte("123456789")
var timeToExpire = time.Now().Add(1 * time.Minute)

func main() {
	http.HandleFunc("/singup", SignUp())
	http.HandleFunc("/singin", SingIn())
	http.HandleFunc("/welcome", Welcome())

	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Panic(err)
	}
}

type Authentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Message string
}

type Claims struct {
	Username string
	jwt.RegisteredClaims
}

func SignUp() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			renderJson(w, http.StatusMethodNotAllowed, Response{"method not allowed"})
			return
		}

		var auth Authentication

		err := json.NewDecoder(r.Body).Decode(&auth)
		if err != nil {
			renderJson(w, http.StatusBadRequest, Response{"inform username and password"})
			return
		}

		if len(auth.Username) == 0 || len(auth.Password) == 0 {
			renderJson(w, http.StatusBadRequest, Response{"inform username and password"})
			return
		}

		if _, ok := DB[auth.Username]; ok {
			renderJson(w, http.StatusBadRequest, Response{"sorry, that username already exist"})
			return
		}

		DB[auth.Username] = auth.Password

		renderJson(w, http.StatusCreated, Response{"successfully"})
	}
}

func SingIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			renderJson(w, http.StatusMethodNotAllowed, Response{"method not allowed"})
			return
		}

		var auth Authentication

		err := json.NewDecoder(r.Body).Decode(&auth)
		if err != nil {
			renderJson(w, http.StatusBadRequest, Response{"inform username and password"})
			return
		}

		if v, ok := DB[auth.Username]; !ok || v != auth.Password {
			renderJson(w, http.StatusUnauthorized, Response{"wrong username or password"})
			return
		}

		claims := Claims{
			Username: auth.Username,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(timeToExpire),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString(JWT_SECRET)
		if err != nil {
			renderJson(w, http.StatusInternalServerError, Response{"the server encountered an error and could not complete your request"})
			return
		}

		renderJson(w, http.StatusOK, map[string]string{
			"token": tokenString,
		})
	}
}

func Welcome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			renderJson(w, http.StatusMethodNotAllowed, Response{"method not allowed"})
			return
		}

		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			renderJson(w, http.StatusUnauthorized, Response{"this route require authentication"})
			return
		}

		tokenString = strings.Split(tokenString, " ")[1]

		var claims = &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				renderJson(w, http.StatusUnauthorized, Response{err.Error()})
				return
			}
			renderJson(w, http.StatusInternalServerError, Response{err.Error()})
			return
		}

		if !token.Valid {
			renderJson(w, http.StatusUnauthorized, Response{"token don't valid"})
			return
		}

		renderJson(w, http.StatusOK, Response{"Welcome!"})
	}
}

func renderJson(w http.ResponseWriter, statusCode int, body interface{}) {
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
	}
}
