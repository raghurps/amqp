package rmq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

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
	addr string
}

// GetRMQClient returns a RMQ client
func GetRMQClient(username, password, url, port, vhost string, secure bool) (*Client, error) {
	connectionType := "amqp"
	if secure {
		connectionType = "amqps"
	}

	addr := fmt.Sprintf("%s://%s:%s@%s:%s%s",
		connectionType,
		username,
		password,
		url,
		port,
		vhost,
	)
	conn, err := amqp.Dial(addr)

	if err != nil {
		log.Println(err.Error())
		return &Client{}, err
	}
	return &Client{conn, addr}, nil
}

func (c *Client) reconnect(retries int, interval time.Duration) (err error) {
	log.Println("Re-connecting to rabbitmq server...")
	count := retries
	for count > 0 {
		count--
		c.conn, err = amqp.Dial(c.addr)
		// return if re-connect succeeded
		if err == nil {
			return
		}

		// Retry if re-connect failed
		log.Println(err.Error())
		if count > 0 {
			log.Printf("Attempt #%d: AMQP connection failed, retrying after %s ...\n", retries-count, interval)
			time.Sleep(interval)
			continue
		}
		return
	}
	return
}

// ChannelOpts ...
type ChannelOpts struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

// DefaultChannelOpts ...
func DefaultChannelOpts() *ChannelOpts {
	return &ChannelOpts{
		PrefetchCount: 1,
		PrefetchSize:  0,
		Global:        false,
	}
}

func (c *Client) getChannel(opts *ChannelOpts) (ch *amqp.Channel, err error) {
	ch, err = c.conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return
	}

	defaultOpts := DefaultChannelOpts()

	if opts != nil {
		defaultOpts = opts
	}

	err = ch.Qos(
		defaultOpts.PrefetchCount, // prefetch count
		defaultOpts.PrefetchSize,  // prefetch size
		defaultOpts.Global,        // global
	)
	if err != nil {
		log.Println(err.Error())
		return
	}
	return
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

	// log.Printf("Publishing message: %s\n\n\n%v\n", string(msg.Body), msg)

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
type SubscribeOpts struct {
	CorrelationID      string        // Correlation ID
	Reconnect          bool          // Reconnect if connection closed
	ReconnectRetries   int           // Number of retries for reconnecting
	ReconnectInterval  time.Duration // Interval to wait before retrying connection
	ListenIndefinitely bool          // Listen indefinitely
	PublishResponse    bool          // Publish response from handler
}

// DefaultSubscribeOpts ...
func DefaultSubscribeOpts() *SubscribeOpts {
	return &SubscribeOpts{
		"",
		false,
		0,
		0 * time.Second,
		false,
		false,
	}
}

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
	opts *SubscribeOpts,
	chanOpts *ChannelOpts,
	handler func(amqp.Delivery) (amqp.Publishing, error),
) error {

	ch, err := c.getChannel(chanOpts)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer ch.Close()

	// Ensure a consumer does not consume another message unless it has processed
	// the last message
	//err = ch.Qos(
	//	1,     // prefetch count
	//	0,     // prefetch size
	//	false, // global
	//)
	//if err != nil {
	//	log.Println(err.Error())
	//	return err
	//}

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
			if opts.Reconnect {
				err = c.reconnect(opts.ReconnectRetries, opts.ReconnectInterval)
				if err != nil {
					return err
				}

				ch, err = c.getChannel(chanOpts)
				if err != nil {
					return err
				}

				msgs, err = ch.Consume(
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

				continue
			}
			return errors.New("connection closed/interrupted")
		}

		select {
		case msg := <-msgs:
			if len(msg.Body) == 0 {
				log.Println("Received empty message. Ignoring...")
				continue
			}

			//log.Printf("Received message: %s\n\n\n%v\n", string(msg.Body), msg)

			if opts.CorrelationID != "" && msg.CorrelationId != opts.CorrelationID {
				log.Printf("Re-queuing message as "+
					"correlationIDs don't match. Got: [%s] Expected: [%s]\n",
					msg.CorrelationId, opts.CorrelationID)
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
				return err
			}

			msg.Ack(false)

			// If subscriber doesn't want to publish response
			// skip the response publishing part
			if opts.PublishResponse {
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
			if opts.ListenIndefinitely {
				continue
			}

			return nil
		case <-ctx.Done():
			return nil
		}
	}
}
