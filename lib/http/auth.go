package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func NewAuthClient() *AuthClient{
	return &AuthClient{
		BaseURL: "http://localhost:8081",
		}
	}

func (c *AuthClient) VerifyToken(Token string) (bool, string, error) {
	requestBody, err := json.Marshal(map[string]string{"Token": Token})
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(c.BaseURL+"/auth/verify", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()



	var result VerifyResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Valid, result.UserName, nil
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	authClient := NewAuthClient()

	return func(w http.ResponseWriter, r *http.Request){
		if r.URL.Path == "/login" {
			next(w,r)
			return
		}
	

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if token == authHeader {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Неверный токен"))
		return
	}

	valid, username, err := authClient.VerifyToken(token)
	if err != nil || !valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Неверный токен"))
		return
	}


	r.Header.Set("X-User-Name", username)
	next(w, r)
	}
}

func loginFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
	authClient := NewAuthClient()

	resp, err := http.Post(authClient.BaseURL+"/auth/login", "application/json", r.Body)
	if err != nil {

		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		w.Write([]byte("Ошибка при логине"))
		return 
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	next(w, r)
	}
}

func regFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		authClient := NewAuthClient()
		
		resp, err := http.Post(authClient.BaseURL + "/register", "application/json", r.Body)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return 
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			w.Write([]byte("Ошибка при регистрации"))
			return 
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		next(w, r)
	}
}
