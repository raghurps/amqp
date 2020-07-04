package rmq

import (
	"log"

	"github.com/streadway/amqp"
)

/*
DeclareQueueOpts ...

Durable and Non-Auto-Deleted queues will survive server restarts and remain when
there are no remaining consumers or bindings. Persistent publishings will be restored
in this queue on server restart. These queues are only able to be bound to durable exchanges.

Non-Durable and Auto-Deleted queues will not be redeclared on server restart and will
be deleted by the server after a short time when the last consumer is canceled or the last
consumer's channel is closed. Queues with this lifetime can also be deleted normally with
QueueDelete. These durable queues can only be bound to non-durable exchanges.

Non-Durable and Non-Auto-Deleted queues will remain declared as long as the server is running
regardless of how many consumers. This lifetime is useful for temporary topologies that may
have long delays between consumer activity. These queues can only be bound to non-durable exchanges.

Durable and Auto-Deleted queues will be restored on server restart, but without active consumers
will not survive and be removed. This Lifetime is unlikely to be useful.

Exclusive queues are only accessible by the connection that declares them and will be deleted
when the connection closes. Channels on other connections will receive an error when attempting
to declare, bind, consume, purge or delete a queue with the same name.

When noWait is true, the queue will assume to be declared on the server. A channel exception
will arrive if the conditions are met for existing queues or attempting to modify an existing
queue from a different connection.
*/
type DeclareQueueOpts struct {
	Durable    bool       // default true
	AutoDelete bool       // default false
	Exclusive  bool       // default false
	NoWait     bool       // default false
	Args       amqp.Table // default nil
}

// DefaultDeclareQueueOpts ...
func DefaultDeclareQueueOpts() *DeclareQueueOpts {
	return &DeclareQueueOpts{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
}

/*
QueueDeclare declares a queue on the RabbitMQ server

name is the name of queue

opts is the options for declaring a queue
*/
func (c *Client) QueueDeclare(name string, opts *DeclareQueueOpts) (amqp.Queue, error) {
	defaultOpts := DefaultDeclareQueueOpts()

	if opts != nil {
		defaultOpts = opts
	}
	var q amqp.Queue
	ch, err := c.conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return q, err
	}
	defer ch.Close()

	q, err = ch.QueueDeclare(
		name,
		defaultOpts.Durable,
		defaultOpts.AutoDelete,
		defaultOpts.Exclusive,
		defaultOpts.NoWait,
		defaultOpts.Args,
	)
	if err != nil {
		log.Println(err.Error())
		return q, err
	}

	return q, nil
}

// QueueBindOpts ...
type QueueBindOpts struct {
	NoWait bool       // default false
	Args   amqp.Table // default nil
}

// DefaultQueueBindOpts ...
func DefaultQueueBindOpts() *QueueBindOpts {
	return &QueueBindOpts{
		NoWait: false,
		Args:   nil,
	}
}

/*
QueueBind binds a queue to an exchange with provided routing key on the RabbitMQ server
*/
func (c *Client) QueueBind(exchange, queue, key string, opts *QueueBindOpts) error {
	defaultOpts := DefaultQueueBindOpts()

	if opts != nil {
		defaultOpts = opts
	}

	ch, err := c.conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer ch.Close()

	err = ch.QueueBind(
		queue,
		key,
		exchange,
		defaultOpts.NoWait,
		defaultOpts.Args,
	)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}
