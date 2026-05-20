package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Kitten-King/user-sdk"
	"github.com/gorilla/mux"
)

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user *user_sdk.User) error
	GetByID(ctx context.Context, id int) (*user_sdk.UserWithCity, error)
	FindWithinRadius(ctx context.Context, lat, lon, radius float64) ([]user_sdk.UserWithCity, error)
}

type KafkaProducerInterface interface {
	PublishUserCreated(user *user_sdk.User) error
}

type UserHandler struct {
	repo          UserRepositoryInterface
	kafkaProducer KafkaProducerInterface
}

func NewUserHandler(repo UserRepositoryInterface, kp KafkaProducerInterface) *UserHandler {
	return &UserHandler{repo: repo, kafkaProducer: kp}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user user_sdk.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 1. Пишем в свою БД
	if err := h.repo.CreateUser(r.Context(), &user); err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// 2. Публикуем событие в Кафку для Гейтвея!
	if h.kafkaProducer != nil {
		_ = h.kafkaProducer.PublishUserCreated(&user)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) FindWithinRadius(w http.ResponseWriter, r *http.Request) {
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	radius, _ := strconv.ParseFloat(r.URL.Query().Get("r"), 64)

	users, err := h.repo.FindWithinRadius(r.Context(), lat, lon, radius)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
