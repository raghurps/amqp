package rmq

import (
	"log"

	"github.com/streadway/amqp"
)

/*
DeclareExchangeOpts ...

Each exchange belongs to one of a set of exchange kinds/types implemented by the server.
The exchange types define the functionality of the exchange - i.e. how messages are routed through it.
Once an exchange is declared, its type cannot be changed.
The common types are "direct", "fanout", "topic" and "headers".

Durable and Non-Auto-Deleted exchanges will survive server restarts
and remain declared when there are no remaining bindings. This is the
best lifetime for long-lived exchange configurations like stable routes and default exchanges.

Non-Durable and Auto-Deleted exchanges will be deleted when there
are no remaining bindings and not restored on server restart.
This lifetime is useful for temporary topologies that should not
pollute the virtual host on failure or after the consumers have completed.

Non-Durable and Non-Auto-deleted exchanges will remain as long as the server
is running including when there are no remaining bindings. This is useful for
temporary topologies that may have long delays between bindings.

Durable and Auto-Deleted exchanges will survive server restarts and will be
removed before and after server restarts when there are no remaining bindings.
These exchanges are useful for robust temporary topologies or when you require
binding durable queues to auto-deleted exchanges.

Note: RabbitMQ declares the default exchange types like 'amq.fanout' as durable,
so queues that bind to these pre-declared exchanges must also be durable.

Exchanges declared as `internal` do not accept accept publishings. Internal
exchanges are useful when you wish to implement inter-exchange topologies that
should not be exposed to users of the broker.

When noWait is true, declare without waiting for a confirmation from the server.
The channel may be closed as a result of an error. Add a NotifyClose listener to
respond to any exceptions.

Optional amqp.Table of arguments that are specific to the server's implementation
of the exchange can be sent for exchange types that require extra parameters.
*/
type DeclareExchangeOpts struct {
	Kind        string     // default amqp.ExchangeDirect
	Durable     bool       // default true
	AutoDeleted bool       // default false
	Internal    bool       // default false
	NoWait      bool       // default false
	Args        amqp.Table // default nil
}

// DefaultDeclareExchangeOpts returns default DeclareExchangeOpts
func DefaultDeclareExchangeOpts() *DeclareExchangeOpts {
	return &DeclareExchangeOpts{
		Kind:        amqp.ExchangeDirect,
		Durable:     true,
		AutoDeleted: false,
		Internal:    false,
		NoWait:      false,
		Args:        nil,
	}
}

/*
ExchangeDeclare declares an exchange on the RabbitMQ server

name is name of the exhange

opts is options for declaring an exchange
*/
func (c *Client) ExchangeDeclare(name string, opts *DeclareExchangeOpts) error {
	defaultOpts := DefaultDeclareExchangeOpts()

	// update defaultOpts if opts provided
	if opts != nil {
		defaultOpts = opts
	}
	ch, err := c.conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		name,                    // name
		defaultOpts.Kind,        // type
		defaultOpts.Durable,     // durable
		defaultOpts.AutoDeleted, // auto-deleted
		defaultOpts.Internal,    // internal
		defaultOpts.NoWait,      // no-wait
		defaultOpts.Args,        // arguments
	)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}
