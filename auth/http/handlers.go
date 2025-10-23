package http

import (
	"auth/database"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type HTTPHandlers struct {
	database database.Postgres
	Logged isLogged
}

func NewHTTPHandlers(db database.Postgres) *HTTPHandlers{
	return &HTTPHandlers{
		database: db,
		Logged: isLogged{},
	}
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func (h *HTTPHandlers) HandleRegUser(w http.ResponseWriter, r *http.Request) {
	var RegDTO = RegDTO{}
		if h.Logged.isLogged{
		w.WriteHeader(http.StatusConflict) // тут редирект на главную страницу
		w.Write([]byte("Вы не можете создать новый аккаунт, так как уже залогинены"))
		return
	} 
	if err := json.NewDecoder(r.Body).Decode(&RegDTO); err != nil{
		errDTO := CreateErrDTO(err.Error(), time.Now())

		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}
	if err := RegDTO.ValidateToCreate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(RegDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("failed to hash password: %w", err)
		return
	}
	if err := h.database.DbRegUser(RegDTO.Username, hashedPass); err != nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Такой аккаунт уже есть"))
		return
	}
	b, err := json.MarshalIndent(RegDTO, "", "	")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(b); err != nil {
		fmt.Print(err)
	}

	if _, err := w.Write([]byte("\nУспешная регистрация")); err != nil {
		fmt.Println("failed to write http response", err)
		return
	}
}


func (h *HTTPHandlers) HandleLogUser(w http.ResponseWriter, r *http.Request) {
	var LogDTO = LogDTO{}
	if h.Logged.isLogged{
		w.Write([]byte("Вы уже залогинены")) // тут редирект на главную страницу
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&LogDTO); err != nil{
		errDTO := CreateErrDTO(err.Error(), time.Now())
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}
	if err := LogDTO.ValidateToCreate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hash, err := h.database.DbLogUser(LogDTO.Username, LogDTO.Password)	
	if err != nil{
		fmt.Print("Ошибка при входе в аккаунт: ", err)
		w.WriteHeader(http.StatusUnauthorized)
		return 
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(LogDTO.Password))
	if err != nil {
		w.Write([]byte("Неверный пароль или логин"))
		w.WriteHeader(http.StatusBadRequest)
		return 
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":		LogDTO.Username,
		"exp":			time.Now().Add(24 *time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwtSecret)

	json.NewEncoder(w).Encode(map[string]string{
    	"token": tokenString,
    	"username": LogDTO.Username,
	})

	h.Logged.isLogged = true
	w.WriteHeader(http.StatusOK)
	
	if _, err := w.Write([]byte("Успешный вход в аккаунт")); err != nil {
		fmt.Println("failed to write http response: ", err)
		return
	}
}

func (h *HTTPHandlers) HandleVerify(w http.ResponseWriter, r *http.Request){
	var req struct {
		Token string `json:"Token"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error)  {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		w.Write([]byte(`{"Valid":false}`))
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)

	jsonData, err := json.Marshal(map[string]string{
		"Valid": "true",
		"UserName": claims["username"].(string),
	})
	if err != nil {
		panic(err)
	}
	w.Write(jsonData)
	
}


func (h *HTTPHandlers) HandleLogout(w http.ResponseWriter, r *http.Request){
	if h.Logged.isLogged{
	h.Logged.isLogged = false
	w.Write([]byte("Успешный выход из аккаунта"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}