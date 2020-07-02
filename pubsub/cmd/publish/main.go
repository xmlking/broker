package main

import (
	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xmlking/broker/pubsub"
)

func main() {
	// broker.DefaultBroker = broker.NewBroker(broker.ProjectID("my-project-id")); // use cfg.pubsub.ProjectID
	broker.DefaultBroker = broker.NewBroker()

	msg := pubsub.Message{
		ID:         uuid.New().String(),
		Data:       []byte("ABC€"),
		Attributes: map[string]string{"sumo": "demo"},
	}

	if err := broker.Publish("ingestion-in-dev", &msg); err != nil {
		log.Error().Err(err).Send()
	}
}
