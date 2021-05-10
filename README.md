[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://raghu-gitpod.platform9.horse/#https://github.com/raghuP9/amqp)

# amqp

Use this library to make rpc calls over rabbit MQ using amqp protocol

## Quickstart

### Download library

```bash
go get -u github.com/raghuP9/amqp
```

### Test with example

#### Generate amqpctl binary

```bash
go get -u github.com/raghuP9/amqp/cmd/amqpctl
$GOPATH/bin/amqpctl --help
```

#### Publish message

```bash
$GOPATH/bin/amqpctl --server rabbitmq.example.com --port 5672 --username <username> --password <password> producer
```

#### Consume message

```bash
$GOPATH/bin/amqpctl --server rabbitmq.example.com --port 5672 --username <username> --password <password> consumer
```

### Use in go code

GoDoc [Link](https://pkg.go.dev/github.com/raghuP9/amqp@v0.0.2/pkg/rpc/rmq)

#### Create client object

```go
package main

import (
  "github.com/raghuP9/amqp/blob/master/pkg/rpc/rmq"
)

func main() {
  client, err := rmq.GetRMQClient(
    true,                       // to connect securely i.e. using amqps or else set to false
    "rabbitmq-username",        // rabbitmq username
    "rabbitmq-password",        // rabbitmq password
    "myrabbitmq.server.com",    // rabbitmq server URL
    "1234",                     // rabbitmq server port
    "/"                         // vhost
  )
}
```

#### Declare exchange

```go
err := client.ExchangeDeclare(
  "exchange-name",
  rmq.DefaultExchangeDeclareOpts(),
  rmq.DefaultConnectOpts(),
  )
```

#### Delete exchange

```go
err := client.ExchangeDelete(
  "exchange-name",  // Exchange name
  true,             // IfUnused: Remove exchange if no queue bound to this exchange
  false,            // NoWait: Do not wait for deletion confirmation from rabbitmq server
)
```

#### Declare queue

```go
err := client.QueueDeclare(
  "queue-name",
  rmq.DefaultDeclareQueueOpts(),
  rmq.DefaultConnectOpts(),
  )
```

#### Bind queue to an exchage using routing key

```go
err := client.QueueBind(
  "exchange-name",
  "queue-name",
  "routing-key",
  rmq.DefaultQueueBindOpts(),
  rmq.DefaultConnectOpts(),
)
```

#### Delete queue

```go
err := client.QueueDelete(
  "queue-name",
  rmq.DefaultQueueDeleteOpts(),
  rmq.DefaultConnectOpts(),
  )
```

#### Purge queue

```go
err := client.QueuePurge(
  "queue-name",
  false,        // NoWait: do not wait for confirmation from rabbitmq server and return
  rmq.DefaultConnectOpts(),
)
```

#### Publish messages

```go
import "github.com/streadway/amqp"

func doSomething() {
  err := client.Publish(
    amqp.Publishing{
      Body:         []byte(c.String("message")),
      DeliveryMode: amqp.Persistent,
      ContentType:  "plain/text",
      Timestamp:    time.Now(),
    },
    "exchange-name",
    "routing-key",
    rmq.DefaultPublishOpts(),
    rmq.DefaultConnectOpts(),
  )
}
```

#### Subscribe to a queue for messages and take actions on different messages

```go
import "github.com/streadway/amqp"

func handler(msg amqp.Delivery) (amqp.Publishing, error) {
...
}

func doSomething() {
  err := client.Subscribe(
    context.TODO(),              // When passing context with cancel func, calling cancel() will return from Subscribe function
    "queue-name",                // Name of the queue
    rmq.DefaultSubscribeOpts(),  // Provide options such as correlation ID, listen indefinitely on the queue, reconnect if disconnected, publish response from handler function
    rmq.DefaultChannelOpts(),    // Set Qos e.g. do not pick another message from queue unless previous message is processed.
    rmq.DefaultConnectOpts(),
    handler,                     // handler function that takes received message, processes it and returns response message, can be anonymous fn.
  )
}
```
