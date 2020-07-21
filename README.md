# amqp
Use this library to make rpc calls over rabbit MQ using amqp protocol

# Quickstart
### Download library 
```
# go get -u github.com/raghuP9/amqp
```
### Test with example
#### Generate amqpctl binary
```
# go get -u github.com/raghuP9/amqp/cmd/amqpctl
# $GOPATH/bin/amqpctl --help
```
#### Publish message
```
# $GOPATH/bin/amqpctl --server rabbitmq.example.com --port 5672 --username <username> --password <password> producer
```
#### Consume message
```
# $GOPATH/bin/amqpctl --server rabbitmq.example.com --port 5672 --username <username> --password <password> consumer
```
