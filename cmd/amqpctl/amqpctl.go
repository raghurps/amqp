package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/raghuP9/amqp/pkg/rpc"
	"github.com/raghuP9/amqp/pkg/rpc/rmq"
	"github.com/streadway/amqp"
	"github.com/urfave/cli/v2"
)

var client *rmq.Client

func bootstrap(c rpc.RabbitMQRPC, queue string) error {

	// Declare queue if it doesn't exist
	_, err := c.QueueDeclare(queue, nil)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func getRMQClient(c *cli.Context) *rmq.Client {
	serverURL := c.String("server")
	username := c.String("user")
	password := c.String("password")
	vhost := c.String("vhost")
	port := fmt.Sprintf("%d", c.Int("port"))
	secure := c.Bool("secure")
	queue := c.String("queue")

	client, err := rmq.GetRMQClient(
		username,
		password,
		serverURL,
		port,
		vhost,
		secure,
	)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	err = bootstrap(client, queue)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return client
}

func beforeFunc(c *cli.Context) error {
	client = getRMQClient(c)
	return nil
}

func afterFunc(c *cli.Context) error {
	return nil
}

func produce(c *cli.Context) error {
	if client == nil {
		err := errors.New("Nil RabbitMQ client")
		log.Println(err.Error())
		return err
	}
	err := client.Publish(
		amqp.Publishing{
			Body:         []byte(c.String("message")),
			DeliveryMode: amqp.Persistent,
			ContentType:  "plain/text",
			Timestamp:    time.Now(),
		},
		"",
		c.String("queue"),
		nil,
	)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	log.Printf("Published message: %s", c.String("message"))
	return nil
}

func consumer(c *cli.Context) error {
	if client == nil {
		err := errors.New("Nil RabbitMQ client")
		log.Println(err.Error())
		return err
	}
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	subsOpt := rmq.DefaultSubscribeOpts()

	chanOpts := rmq.DefaultChannelOpts()
	err := client.Subscribe(
		ctx,
		c.String("queue"),
		subsOpt,
		chanOpts,
		func(msg amqp.Delivery) (amqp.Publishing, error) {
			log.Printf("Received message: %s", string(msg.Body))
			return amqp.Publishing{}, nil
		},
	)

	if err != nil {
		log.Fatalf("Exiting due to error: %s\n", err.Error())
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:    "amqpctl",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "server",
				Usage:    "provide rabbitMQ server url",
				Value:    "127.0.0.1",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "port",
				Usage:    "provide rabbitMQ server url",
				Required: true,
				Value:    8443,
			},
			&cli.BoolFlag{
				Name:  "secure",
				Usage: "use amqps instead of amqp",
			},
			&cli.StringFlag{
				Name:  "vhost",
				Usage: "Provide vhost",
				Value: "/",
			},
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"username"},
				Usage:    "Provide rabbitMQ username",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Usage:    "Provide rabbitMQ password",
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "producer",
				Usage: "Produces message on rabbitMQ exchange",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "queue",
						Usage: "Provide queue to publish message",
						Value: "testamqp",
					},
					&cli.StringFlag{
						Name:  "message",
						Usage: "Provide message text to send over rabbitMQ",
						Value: "Hello World!",
					},
				},
				Before: beforeFunc,
				Action: produce,
				After:  afterFunc,
			},
			{
				Name:  "consumer",
				Usage: "Consumes message from rabbitMQ exchange",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "queue",
						Usage: "Provide queue to read messages from",
						Value: "testamqp",
					},
				},
				Before: beforeFunc,
				Action: consumer,
				After:  afterFunc,
			},
		},
	}
	app.Run(os.Args)
}
