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

type amqpConsumeFx = func(msg amqp091.Delivery) error

type amqp struct {
	connMu *sync.RWMutex
	conn   *amqp091.Connection

	channelMu *sync.RWMutex
	channel   map[string]*amqp091.Channel

	queueMu *sync.RWMutex
	queue   map[string]struct {
		// _    amqp091.Queue
		opts amqpQueueOpts
	}

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
func (a *amqp) Start(url string, initReady chan<- struct{}) {
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

		if isFirstConnection {
			initReady <- struct{}{}
			close(initReady)
		}

		a.connMu.Lock()
		a.conn = conn
		a.closeNotifier = a.conn.NotifyClose(make(chan *amqp091.Error))
		a.connMu.Unlock()

		isFirstConnection = false
		backoff = time.Second
		a.restore()

		a.errChan <- fmt.Errorf("utility<amqp.Connect>: %w", <-a.closeNotifier)
	}
}

func (a *amqp) restore() error {
	var tempChannel string
	for channel, _ := range a.channel {
		if err := a.AddChannel(channel); err != nil {
			return fmt.Errorf("utility<amqp.restore>: %w", err)
		}
		tempChannel = channel
	}

	for _, spec := range a.queue {
		if err := a.AddQueue(tempChannel, spec.opts); err != nil {
			return fmt.Errorf("utility<amqp.restore>: %w", err)
		}
	}

	for _, spec := range a.consumer {
		if err := a.AddConsumer(tempChannel, spec.fx, spec.opts); err != nil {
			return fmt.Errorf("utility<amqp.restore>: %w", err)
		}
	}
	return nil
}

// Starts monitor on activites indefinitely.
func (a amqp) Monitor(end <-chan struct{}) {
	go func() {
		for {
			select {
			case err := <-a.errChan:
				fmt.Println(fmt.Errorf("utility<amqp.Monitor>: %w", err))
			case <-end:
				return
			}
		}
	}()
}

// Registers a channel to be managed. If done multiple times
// using the same name, the old one would be replaced
func (a *amqp) AddChannel(name string) error {
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
func (a *amqp) AddQueue(channel string, opts amqpQueueOpts) error {
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
	a.queue[opts.name] = struct {
		// _    amqp091.Queue
		opts amqpQueueOpts
	}{opts}
	a.queueMu.Unlock()

	return nil
}

// Registers a consumer to be managed and starts a goroutine for consuming. If
// done multiple times, the old one would be replaced.
//
// The consumer ACKs the message manually. If fx has an error, messages would be
// NACK-ed
func (a amqp) AddConsumer(
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
	if oldConsumer, ok := a.consumer[channel]; ok {
		oldConsumer.end()
	}
	a.consumer[channel] = struct {
		end  context.CancelFunc
		fx   amqpConsumeFx
		opts amqpConsumeOpts
	}{cancel, fx, opts}
	a.consumerMu.Unlock()

	go func() {
		for msg := range msgs {
			err := fx(msg)
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

func (a amqp) Publish(channel string, payload []byte, opts amqpPublishOpts) error {
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

// Wraps call to Channel.QueueBind
func (a amqp) QueueBind(channel string, opts amqpQueueBindOpts) error {
	a.channelMu.RLock()
	c, cOk := a.channel[channel]
	if !cOk {
		err := fmt.Errorf("Channel %s is not registered yet", channel)
		return fmt.Errorf("utility<amqp.Consume>: %w", err)
	}
	a.channelMu.RUnlock()

	return c.QueueBind(
		opts.name,
		opts.key,
		opts.exchange,
		opts.noWait,
		opts.args)
}

func NewAmqp() amqp {
	return amqp{
		connMu: &sync.RWMutex{},

		queueMu: &sync.RWMutex{},
		queue: make(map[string]struct {
			// _    amqp091.Queue
			opts amqpQueueOpts
		}),
		channelMu: &sync.RWMutex{},
		channel:   make(map[string]*amqp091.Channel),

		consumerMu: &sync.RWMutex{},
		consumer: make(map[string]struct {
			end  context.CancelFunc
			fx   amqpConsumeFx
			opts amqpConsumeOpts
		}),

		errChan: make(chan error, 20)}
}
