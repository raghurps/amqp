package rpc

import (
	"context"

	"github.com/raghuP9/amqp/pkg/rpc/rmq"
	"github.com/streadway/amqp"
)

// RabbitMQRPC ...
type RabbitMQRPC interface {
	ExchangeDeclare(string, *rmq.DeclareExchangeOpts, *rmq.ConnectOpts) error
	ExchangeDelete(string, bool, bool, *rmq.ConnectOpts) error
	QueueDeclare(string, *rmq.DeclareQueueOpts, *rmq.ConnectOpts) (amqp.Queue, error)
	QueueBind(string, string, string, *rmq.QueueBindOpts, *rmq.ConnectOpts) error
	QueuePurge(string, bool, *rmq.ConnectOpts) error
	QueueDelete(string, *rmq.QueueDeleteOpts, *rmq.ConnectOpts) error
	Publish(amqp.Publishing, string, string, *rmq.PublishOpts, *rmq.ConnectOpts) error
	Subscribe(
		context.Context,
		string,
		*rmq.SubscribeOpts,
		*rmq.ChannelOpts,
		*rmq.ConnectOpts,
		func(amqp.Delivery) (amqp.Publishing, error),
	) error
}
