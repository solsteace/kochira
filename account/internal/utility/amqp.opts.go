package utility

import "github.com/rabbitmq/amqp091-go"

type amqpQueueOpts struct {
	name       string        // what should we name the queue we're going to use?
	durable    bool          // [read QueueDeclare doc, for now]
	autoDelete bool          // [read QueueDeclare doc, for now]
	exclusive  bool          // can this queue be accessed using other connection (channel, I guess)?
	noWait     bool          // should we assume this queue had already been declared on the broker?
	opts       amqp091.Table // arguments
}

func NewDefaultAmqpQueueOpts(name string) amqpQueueOpts {
	return amqpQueueOpts{
		name:       name,
		durable:    false,
		autoDelete: false,
		exclusive:  false,
		noWait:     false,
		opts:       nil}
}

type amqpConsumeOpts struct {
	queue     string        // which queue we're receiving messages from?
	consumer  string        // what name should I identify myself with?
	autoAck   bool          // upon receiving message, should it be ACK-ed automatically?
	exclusive bool          // [read channel.Consume doc, for now]
	noLocal   bool          // [read channel.Consume doc, for now]
	noWait    bool          // [read channel.Consume doc, for now]
	opts      amqp091.Table // arguments

	requeueNack bool // Should the consumed message be requeued upon being NACK-ed?
}

func NewDefaultAmqpConsumeOpts(queue string, requeueOnNack bool) amqpConsumeOpts {
	return amqpConsumeOpts{
		queue:       queue,
		consumer:    "",
		autoAck:     false,
		exclusive:   false,
		noLocal:     false,
		noWait:      false,
		opts:        nil,
		requeueNack: requeueOnNack}
}

type amqpPublishOpts struct {
	exchange  string // which exchange to send the message to?
	key       string // what routing key to use?
	mandatory bool   // [read channel.Publish doc, for now]
	immediate bool   // [read channel.Publish doc, for now]

	publishingContentType string
}

func NewDefaultAmqpPublishOpts(
	exchange string,
	key string,
	contentType string,
) amqpPublishOpts {
	return amqpPublishOpts{
		exchange:              exchange,
		key:                   key,
		mandatory:             false,
		immediate:             false,
		publishingContentType: contentType}
}

type amqpQueueBindOpts struct {
	name     string        // Which queue to bind?
	key      string        // What key the publisher should use to send a message to this queue?
	exchange string        // Which exchange to be bound with the queue?
	noWait   bool          // [read channel.QueueBind, for now]
	args     amqp091.Table // args
}

func NewDefaultAmqpQueueBindOpts(name, key, exchange string) amqpQueueBindOpts {
	return amqpQueueBindOpts{
		name:     name,
		key:      key,
		exchange: exchange,
		noWait:   false,
		args:     nil}
}
