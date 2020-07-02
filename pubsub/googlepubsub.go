package broker

import (
	"context"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pubsubBroker struct {
	client  *pubsub.Client
	options Options
	subs    []*subscriber
}

// A pubsub subscriber that manages handling of messages
type subscriber struct {
	options SubscribeOptions
	topic   string
	exit    chan bool
	sub     *pubsub.Subscription
}

func (s *subscriber) run(hdlr Handler) {
	if s.options.Context != nil {
		if max, ok := s.options.Context.Value(maxOutstandingMessagesKey{}).(int); ok {
			s.sub.ReceiveSettings.MaxOutstandingMessages = max
		}
		if max, ok := s.options.Context.Value(maxExtensionKey{}).(time.Duration); ok {
			s.sub.ReceiveSettings.MaxExtension = max
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	for {
		select {
		case <-s.exit:
			cancel()
			return
		default:
			if err := s.sub.Receive(ctx, func(ctx context.Context, pm *pubsub.Message) {
				// If the error is nil lets check if we should auto ack
				err := hdlr(ctx, pm)
				if err == nil {
					// auto ack?
					if s.options.AutoAck {
						pm.Ack()
					}
				}
			}); err != nil {
				log.Error().Err(err).Msg("Receive Error")
				time.Sleep(time.Second)
				continue
			}
		}
	}
}

func (s *subscriber) Options() SubscribeOptions {
	return s.options
}

func (s *subscriber) Topic() string {
	return s.topic
}

func (s *subscriber) Unsubscribe() error {
	select {
	case <-s.exit:
		return nil
	default:
		close(s.exit)
		if deleteSubscription, ok := s.options.Context.Value(deleteSubscription{}).(bool); !ok || deleteSubscription {
			return s.sub.Delete(context.Background())
		}
		return nil
	}
}

func (b *pubsubBroker) Connect() error {
	return nil
}

// Shutdown shuts down all subscribers gracefully and then close the connection
func (b *pubsubBroker) Shutdown() (err error) {
	// close all subs and then connection.
	for _, sub := range b.subs {
		log.Info().Msgf("Unsubscribing from topic: %s", sub.Topic())
		if err = sub.Unsubscribe(); err != nil {
			return
		}
	}
	return b.client.Close()
}

func (b *pubsubBroker) Options() Options {
	return b.options
}

// Publish checks if the topic exists and then publishes via google pubsub
func (b *pubsubBroker) Publish(topic string, msg *pubsub.Message, opts ...PublishOption) (err error) {
	t := b.client.Topic(topic)
	ctx := context.Background()

	pr := t.Publish(ctx, msg)
	if _, err = pr.Get(ctx); err != nil {
		// create Topic if not exists
		if status.Code(err) == codes.NotFound {
			log.Info().Msgf("Topic not exists. creating Topic: %s", topic)
			if t, err = b.client.CreateTopic(ctx, topic); err == nil {
				_, err = t.Publish(ctx, msg).Get(ctx)
			}
		}
	}
	return
}

// Subscribe registers a subscription to the given topic against the google pubsub api
func (b *pubsubBroker) Subscribe(topic string, h Handler, opts ...SubscribeOption) error {
	options := SubscribeOptions{
		AutoAck: true,
		Queue:   "q-" + uuid.New().String(),
		Context: b.options.Context,
	}

	for _, o := range opts {
		o(&options)
	}

	ctx := context.Background()
	sub := b.client.Subscription(options.Queue)

	if createSubscription, ok := b.options.Context.Value(createSubscription{}).(bool); !ok || createSubscription {
		exists, err := sub.Exists(ctx)
		if err != nil {
			return err
		}

		if !exists {
			tt := b.client.Topic(topic)
			subb, err := b.client.CreateSubscription(ctx, options.Queue, pubsub.SubscriptionConfig{
				Topic:       tt,
				AckDeadline: time.Duration(0),
			})
			if err != nil {
				return err
			}
			sub = subb
		}
	}

	subscriber := &subscriber{
		options: options,
		topic:   topic,
		exit:    make(chan bool),
		sub:     sub,
	}

	// keep track of subs
	b.subs = append(b.subs, subscriber)

	go subscriber.run(h)

	return nil
}

func (b *pubsubBroker) String() string {
	return "googlepubsub"
}

// NewBroker creates a new google pubsub broker
func NewBroker(opts ...Option) Broker {
	options := Options{
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	// retrieve project id
	prjID, _ := options.Context.Value(projectIDKey{}).(string)

	// if `GOOGLE_CLOUD_PROJECT_ID` is present, it will overwrite programmatically set projectID
	if envPrjID := os.Getenv("GOOGLE_CLOUD_PROJECT_ID"); len(envPrjID) > 0 {
		prjID = envPrjID
	}

	// retrieve client opts
	cOpts, _ := options.Context.Value(clientOptionKey{}).([]option.ClientOption)

	// create pubsub client
	c, err := pubsub.NewClient(context.Background(), prjID, cOpts...)
	if err != nil {
		panic(err.Error())
	}

	return &pubsubBroker{
		client:  c,
		options: options,
	}
}
