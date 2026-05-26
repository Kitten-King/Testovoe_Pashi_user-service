package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"Testovoe_Pashi_user-service/internal/db"
	userHttp "Testovoe_Pashi_user-service/internal/delivery/http"
	"Testovoe_Pashi_user-service/internal/kafka"
	"Testovoe_Pashi_user-service/internal/repository"

	"github.com/gorilla/mux"
	kafkago "github.com/segmentio/kafka-go"
)

type TripCompletedEvent struct {
	UserID            int `json:"user_id"`
	DestinationCityID int `json:"destination_city_id"`
}

func StartKafkaConsumer(userRepo *repository.UserRepository) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "trip-completed-events",
		GroupID:  "user-service-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("Kafka Consumer started, waiting for events...")

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Kafka Consumer error: %v", err)
			continue
		}

		var event TripCompletedEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Kafka: failed to unmarshal message: %v", err)
			continue
		}

		log.Printf("Kafka: Received event! Updating user %d to city %d", event.UserID, event.DestinationCityID)

		err = userRepo.UpdateUserCity(context.Background(), event.UserID, event.DestinationCityID)
		if err != nil {
			log.Printf("Failed to update user city in DB: %v", err)
		} else {
			log.Printf("Successfully updated city for user %d", event.UserID)
		}
	}
}

func main() {
	log.Println("Starting User Service...")

	database := db.Connect()
	defer database.Close()

	repo := repository.NewUserRepository(database)

	go StartKafkaConsumer(repo)

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
