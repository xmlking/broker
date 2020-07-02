# Broker
Framework for building async µServices 

## Usage

### Publisher

```go
package main

import (
	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xmlking/grpc-starter-kit/xmlking/broker/pubsub"
)


func main() {
	broker.DefaultBroker = broker.NewBroker()

	msg := pubsub.Message{
		ID:         uuid.New().String(),
		Data:       []byte("ABC€"),
		Attributes: map[string]string{"sumo": "demo"},
	}

	if err := broker.Publish("my-topic", &msg); err != nil {
		log.Error().Err(err).Send()
	}
}
```

### Subscriber

```go
package main

import (
	"context"
	"os"
	"os/signal"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"

	"github.com/xmlking/grpc-starter-kit/xmlking/broker/pubsub"
)

func main() {
	broker.DefaultBroker = broker.NewBroker()

	myHandler := func(ctx context.Context, msg *pubsub.Message) error {

		log.Info().Interface("event.Message.ID", msg.ID).Send()
		log.Info().Interface("event.Message.Attributes", msg.Attributes).Send()
		log.Info().Interface("event.Message.Data", msg.Data).Send()

		log.Info().Interface("event.Message", msg).Send()
		msg.Ack() // or msg.Nack() // or return error for autoAck
		return nil
	}

	err := broker.Subscribe("my-topic", myHandler, broker.Queue("my-topic-sub"))
	if err != nil {
		log.Error().Err(err).Msg("Failed subscribing to Topic: my-topic")
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Info().Msg("Got to Go...")
	// close all subs and then connection.
	if err := broker.Shutdown(); err != nil {
		log.Fatal().Err(err).Msg("Unexpected disconnect error")
	}
}
```


## Reference 
- https://github.com/nytimes/gizmo/blob/master/pubsub/gcp/gcp.go
- https://github.com/lileio/pubsub/blob/master/providers/google/google.go
- https://github.com/micro/go-plugins/tree/master/broker/googlepubsub
- https://github.com/cloudevents/sdk-go/blob/master/protocol/pubsub/v2/options.go
