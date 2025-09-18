package controller

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/solsteace/kochira/subscription/internal/service"
)

type Perk struct {
	service service.Perk
}

func (p Perk) Infer(msg amqp091.Delivery) error {
	if err := p.service.Infer(1); err != nil {
		return fmt.Errorf("controller<Perk.Infer>: %w", err)
	}
	return nil
}

func NewPerk(service service.Perk) Perk {
	return Perk{service}
}
