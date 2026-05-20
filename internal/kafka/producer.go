package kafka

import (
	"log"

	"github.com/Kitten-King/user-sdk"
)

type DummyProducer struct{}

func NewDummyProducer() *DummyProducer {
	return &DummyProducer{}
}

func (p *DummyProducer) PublishUserCreated(user *user_sdk.User) error {
	log.Printf("[Kafka СИМУЛЯЦИЯ]: Отправлено событие UserCreated для ID %d в топик 'user-events'", user.UserID)
	return nil
}
