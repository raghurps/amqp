package rmq

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

/*
// ClientPool ...
type ClientPool struct {
	pool     []*Client
	lock     *[]sync.Mutex
	poolSize int
}
*/

// Client is rabbitmq client object
type Client struct {
	conn *amqp.Connection
}

// GetRMQClient returns a RMQ client
func GetRMQClient(username, password, url, port, vhost string, secure bool) (*Client, error) {
	connectionType := "amqp"
	if secure {
		connectionType = "amqps"
	}
	conn, err := amqp.Dial(
		fmt.Sprintf("%s://%s:%s@%s:%s%s",
			connectionType,
			username,
			password,
			url,
			port,
			vhost,
		),
	)
	if err != nil {
		log.Println(err.Error())
		return &Client{}, err
	}
	return &Client{conn}, nil
}

// PublishOpts ...
type PublishOpts struct {
	Mandatory bool // default false
	Immediate bool // default false
}

// DefaultPublishOpts ...
func DefaultPublishOpts() *PublishOpts {
	return &PublishOpts{
		Mandatory: false,
		Immediate: false,
	}
}

/*
Publish publishes a message to the exchange

msg is the message that needs to be published on the exchange

exchange is the name of exchange where this message will be published

key is the routing key that will be used for routing the message on exchange
to different queues

opts is option for publishing a message
*/
func (c *Client) Publish(msg amqp.Publishing, exchange, key string, opts *PublishOpts) error {
	defaultOpts := DefaultPublishOpts()

	if opts != nil {
		defaultOpts = opts
	}

	ch, err := c.conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer ch.Close()

	log.Printf("Publishing message: %s\n\n\n%v\n", string(msg.Body), msg)

	err = ch.Publish(
		exchange,
		key,
		defaultOpts.Mandatory,
		defaultOpts.Immediate,
		msg,
	)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

// SubscribeOpts ...
type SubscribeOpts struct{}

/*
Subscribe subscribes you to receive messages from a queue.
It processes one message at a time and responds back with a message
if required. You can subscribe to a queue indefinitely in case
you want to keep on processing new messages.

ctx is the context object that can be used for signaling ctx.Done()

queue is the name of the queue from it will receive messages

corrID is the correlation ID that can be used for filtering specific
messages that can be a reply to previously published message. If empty
string is provided, it will process all the messages. If it is not empty
and corrID doesn't match the message's CorrelationId, the message is not
acknowledged and is re-queued.

indefinitely is a flag that can be used to specify whether the subscriber
should be listening to the messages on the queue indefinitely or return
just after processing one message.

publishResponse is a flag that can be used to specify whether the response
from the handler function should be published on the message's replyTo
key on the exchange.

handler is a function that will process the incoming messages and it should
return response(optional, see publishResponse flag defn) and error object.
*/
func (c *Client) Subscribe(
	ctx context.Context,
	queue string,
	corrID string,
	indefinitely bool,
	publishResponse bool,
	handler func(amqp.Delivery) (amqp.Publishing, error),
) error {

	ch, err := c.conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer ch.Close()

	// Ensure a consumer does not consume another message unless it has processed
	// the last message
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	msgs, err := ch.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	for {
		// Need to check if connection is closed or else
		// msgs channel starts dumping empty messages
		// overwhelming the select clause
		if c.conn.IsClosed() {
			log.Println("Connection closed/interrupted...")
			return errors.New("connection closed/interrupted")
		}

		select {
		case msg := <-msgs:
			if len(msg.Body) == 0 {
				log.Println("Received empty message. Ignoring...")
				continue
			}

			log.Printf("Received message: %s\n\n\n%v\n", string(msg.Body), msg)

			if corrID != "" && msg.CorrelationId != corrID {
				log.Printf("Re-queuing message as "+
					"correlationIDs don't match. Got: [%s] Expected: [%s]\n",
					msg.CorrelationId, corrID)
				msg.Nack(false, true)
				continue
			}

			// call handler to process message
			resp, err := handler(msg)
			if err != nil {
				log.Println(err.Error())
				// requeue if error happened
				// while processing request msg
				msg.Nack(false, true)
			}

			msg.Ack(false)

			// If subscriber doesn't want to publish response
			// skip the response publishing part
			if publishResponse {
				err = ch.Publish(
					msg.Exchange,
					msg.ReplyTo,
					false,
					false,
					resp,
				)
				if err != nil {
					log.Println(err.Error())
					return err
				}
			}

			// Listen indefinitely
			// if requested
			if indefinitely {
				continue
			}

			return nil
		case <-ctx.Done():
			return nil
		}
	}
}
