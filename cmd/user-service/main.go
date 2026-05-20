package main

import (
	"log"
	"net/http"

	"Testovoe_Pashi_user-service/internal/db"
	userHttp "Testovoe_Pashi_user-service/internal/delivery/http"
	"Testovoe_Pashi_user-service/internal/kafka"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting User Service...")

	database := db.Connect()
	defer database.Close()

	repo := NewUserRepository(database)

	kafkaProducer := kafka.NewDummyProducer()

	handler := userHttp.NewUserHandler(repo, kafkaProducer)

	r := mux.NewRouter()
	r.HandleFunc("/users", handler.CreateUser).Methods("POST")
	r.HandleFunc("/users/{id}", handler.GetUserByID).Methods("GET")
	r.HandleFunc("/users/search/radius", handler.FindWithinRadius).Methods("GET")

	log.Println("User Service is running on port :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
