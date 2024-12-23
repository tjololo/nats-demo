package subscriber

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

// New event messages are broadcast to all registered client connection channels
type ClientChan chan string

type ClientConfig struct {
	NatsServerURL string
	NatsSubscription string
}

func Subscriber(config ClientConfig) {
	nc, err := nats.Connect(config.NatsServerURL)
	if err != nil {
		log.Fatalf("Failed to connect to nats server %s\n", err)
	}
	ch := make(chan *nats.Msg, 64)
	sub, err := nc.ChanSubscribe(config.NatsSubscription, ch)
	if err != nil {
		log.Fatalf("Failed to subscribe to %s. Due to %s\n", config.NatsSubscription, err)
	}
	defer sub.Unsubscribe()
	defer close(ch)
	for msg := range ch {
		log.Printf("Message received. Subject: %s Data: %s", msg.Subject, msg.Data)
		if msg.Reply != "" {
			err = msg.Respond([]byte(fmt.Sprintf("Replying to %s on %s", msg.Data, msg.Reply)))
			if err != nil {
				log.Printf("Unable to respond to message %s", err)
			}
			err = msg.Ack()
			if err != nil {
				log.Printf("Unable to ack package %s", err)
			}
		}
	}
}