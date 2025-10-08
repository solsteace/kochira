package utility

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type amqpConsumeFx = func(body []byte) error

type Amqp struct {
	connMu *sync.RWMutex
	conn   *amqp091.Connection

	channelMu *sync.RWMutex
	channel   map[string]*amqp091.Channel

	queueMu *sync.RWMutex
	queue   map[string]amqpQueueOpts

	exchangeMu *sync.RWMutex
	exchange   map[string]amqpExchangeOpts

	bindingMu *sync.RWMutex
	binding   map[string]amqpQueueBindOpts

	consumerMu *sync.RWMutex
	consumer   map[string]struct {
		end  context.CancelFunc
		fx   amqpConsumeFx
		opts amqpConsumeOpts
	}

	closeNotifier chan *amqp091.Error
	errChan       chan error
}

// Starts connecting to a node and restores it when went down. `InitReady` channel
// indicates whether the initial connection has succeeded and would be closed
// upon succession
func (a *Amqp) Start(url string, initReady chan<- struct{}) {
	// Monitors connection indefinitely
	go func() {
		for err := range a.errChan {
			fmt.Println(fmt.Errorf("utility<amqp.Start>: monitor: %w", err))
		}
	}()

	isFirstConnection := true
	for backoff := time.Second; ; {
		if !isFirstConnection {
			fmt.Printf("Reconnecting in %.2fs...\n", backoff.Seconds())
			time.Sleep(backoff)
		}

		conn, err := amqp091.Dial(url)
		if err != nil {
			if backoff < 16*time.Second {
				backoff *= 2
			}
			a.errChan <- fmt.Errorf("utility<amqp.Connect>: %w", err)
			continue
		}

		a.connMu.Lock()
		a.conn = conn
		a.closeNotifier = a.conn.NotifyClose(make(chan *amqp091.Error))
		a.connMu.Unlock()

		if isFirstConnection {
			initReady <- struct{}{}
			close(initReady)
		}

		isFirstConnection = false
		backoff = time.Second
		a.restore()

		a.errChan <- fmt.Errorf("utility<amqp.Connect>: %w", <-a.closeNotifier)
	}
}

func (a *Amqp) restore() error {
	var tempChannel string
	for channel, _ := range a.channel {
		if err := a.AddChannel(channel); err != nil {
			return fmt.Errorf("utility<amqp.restore>: %w", err)
		}
		tempChannel = channel
	}

	for _, opts := range a.exchange {
		if err := a.AddExchange(tempChannel, opts); err != nil {
			return fmt.Errorf("utility<amqp.restore>: %w", err)
		}
	}

	for _, opts := range a.queue {
		if err := a.AddQueue(tempChannel, opts); err != nil {
			return fmt.Errorf("utility<amqp.restore>: %w", err)
		}
	}

	for _, opts := range a.binding {
		if err := a.BindQueue(tempChannel, opts); err != nil {
			return fmt.Errorf("utility<amqp.restore>: %w", err)
		}
	}

	for _, opts := range a.consumer {
		if err := a.AddConsumer(tempChannel, opts.fx, opts.opts); err != nil {
			return fmt.Errorf("utility<amqp.restore>: %w", err)
		}
	}
	return nil
}

// Registers a channel to be managed. If done multiple times
// using the same name, the old one would be replaced
func (a *Amqp) AddChannel(name string) error {
	a.connMu.RLock()
	if a.conn == nil {
		err := errors.New("connection is uninitiated")
		return fmt.Errorf("utility<amqp.RegisterChannel>: %w", err)
	}

	ch, err := a.conn.Channel()
	if err != nil {
		return fmt.Errorf("utility<amqp.RegisterChannel>: %w", err)
	}
	a.connMu.RUnlock()

	a.channelMu.Lock()
	a.channel[name] = ch
	a.channelMu.Unlock()

	return nil
}

// Registers a queue to be managed. If done multiple times using the same name,
// the old one would be replaced
func (a *Amqp) AddQueue(channel string, opts amqpQueueOpts) error {
	a.channelMu.RLock()
	c, cOk := a.channel[channel]
	if !cOk {
		err := fmt.Errorf("Channel %s is not registered yet", channel)
		return fmt.Errorf("utility<amqp.RegisterQueue>: %w", err)
	}
	a.channelMu.RUnlock()

	_, err := c.QueueDeclare(
		opts.name,
		opts.durable,
		opts.autoDelete,
		opts.exclusive,
		opts.noWait,
		opts.opts)
	if err != nil {
		return fmt.Errorf("utility<amqp.RegisterQueue>: %w", err)
	}

	a.queueMu.Lock()
	a.queue[opts.name] = opts
	a.queueMu.Unlock()

	return nil
}

func (a Amqp) AddExchange(channel string, opts amqpExchangeOpts) error {
	a.channelMu.RLock()
	c, cOk := a.channel[channel]
	if !cOk {
		err := fmt.Errorf("Channel %s is not registered yet", channel)
		return fmt.Errorf("utility<amqp.AddExcahnge>: %w", err)
	}
	a.channelMu.RUnlock()

	err := c.ExchangeDeclare(
		opts.name,
		opts.kind,
		opts.durable,
		opts.autoDelete,
		opts.internal,
		opts.noWait,
		opts.opts)
	if err != nil {
		return fmt.Errorf("utility<amqp.AddExchange>: %w", err)
	}

	a.exchangeMu.Lock()
	a.exchange[opts.name] = opts
	a.exchangeMu.Unlock()
	return nil
}

// Registers a consumer to be managed and starts a goroutine for consuming. If
// done multiple times, the old one would be replaced.
//
// The consumer ACKs the message manually. If fx has an error, messages would be
// NACK-ed
func (a *Amqp) AddConsumer(
	channel string,
	fx amqpConsumeFx,
	opts amqpConsumeOpts,
) error {
	a.channelMu.RLock()
	c, cOk := a.channel[channel]
	if !cOk {
		err := fmt.Errorf("Channel %s is not registered yet", channel)
		return fmt.Errorf("utility<amqp.Consume>: %w", err)
	}
	a.channelMu.RUnlock()

	ctx, cancel := context.WithCancel(context.Background())
	msgs, err := c.ConsumeWithContext(
		ctx,
		opts.queue,
		opts.consumer,
		opts.autoAck,
		opts.exclusive,
		opts.noLocal,
		opts.noWait,
		opts.opts)
	if err != nil {
		return fmt.Errorf("utility<amqp.Consume>: %w", err)
	}

	a.consumerMu.Lock()
	if oldConsumer, ok := a.consumer[opts.queue]; ok {
		oldConsumer.end()
	}
	a.consumer[opts.queue] = struct {
		end  context.CancelFunc
		fx   amqpConsumeFx
		opts amqpConsumeOpts
	}{cancel, fx, opts}
	a.consumerMu.Unlock()

	go func() {
		for msg := range msgs {
			err := fx(msg.Body)
			if err != nil {
				log.Printf("Failed to process message: %v\n", err)
				if err := msg.Nack(false, opts.requeueNack); err != nil {
					log.Printf("Failed to NACK message: %v\n", err)
				}
				continue
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("Failed to ACK message: %v\n", err)
			}
		}
	}()
	return nil
}

func (a Amqp) Publish(channel string, payload []byte, opts amqpPublishOpts) error {
	a.channelMu.RLock()
	c, cOk := a.channel[channel]
	if !cOk {
		err := fmt.Errorf("Channel %s is not registered yet", channel)
		return fmt.Errorf("utility<amqp.Consume>: %w", err)
	}
	a.channelMu.RUnlock()

	return c.Publish(
		opts.exchange,
		opts.key,
		opts.mandatory,
		opts.immediate,
		amqp091.Publishing{
			ContentType: opts.publishingContentType,
			Body:        payload})
}

// Wraps call to Channel.BindQueue
func (a Amqp) BindQueue(channel string, opts amqpQueueBindOpts) error {
	a.channelMu.RLock()
	c, cOk := a.channel[channel]
	if !cOk {
		err := fmt.Errorf("Channel %s is not registered yet", channel)
		return fmt.Errorf("utility<amqp.Consume>: %w", err)
	}
	a.channelMu.RUnlock()

	err := c.QueueBind(
		opts.name,
		opts.key,
		opts.exchange,
		opts.noWait,
		opts.args)
	if err != nil {
		return fmt.Errorf("utility<amqp.QueueBind>: %w", err)
	}

	a.bindingMu.Lock()
	a.binding[opts.exchange] = opts
	a.bindingMu.Unlock()
	return nil
}

func NewAmqp() Amqp {
	return Amqp{
		connMu: &sync.RWMutex{},

		queueMu:    &sync.RWMutex{},
		queue:      make(map[string]amqpQueueOpts),
		channelMu:  &sync.RWMutex{},
		channel:    make(map[string]*amqp091.Channel),
		exchangeMu: &sync.RWMutex{},
		exchange:   make(map[string]amqpExchangeOpts),
		bindingMu:  &sync.RWMutex{},
		binding:    make(map[string]amqpQueueBindOpts),

		consumerMu: &sync.RWMutex{},
		consumer: make(map[string]struct {
			end  context.CancelFunc
			fx   amqpConsumeFx
			opts amqpConsumeOpts
		}),

		errChan: make(chan error, 20)}
}
