package service

import (
	"fmt"

	"github.com/solsteace/kochira/account/internal/domain/account"
	"github.com/solsteace/kochira/account/internal/domain/account/messaging"
	accountStore "github.com/solsteace/kochira/account/internal/domain/account/store"
	"github.com/solsteace/kochira/account/internal/utility"
	"github.com/solsteace/kochira/account/internal/utility/hash"
)

const CreateSubscriptionQueue = "subscription.creator" // `depends on `subscription` service

type Account struct {
	accountStore accountStore.User
	hasher       hash.Handler
	messenger    *utility.Amqp // Change to interface later
}

func NewAccount(
	store accountStore.User,
	hasher hash.Handler,
	messenger *utility.Amqp,
) Account {
	return Account{
		accountStore: store,
		hasher:       hasher,
		messenger:    messenger}
}

func (as Account) Register(username, password, email string) error {
	digest, err := as.hasher.Generate(password)
	if err != nil {
		return fmt.Errorf("service<Account.Register>: %w", err)
	}

	user, err := account.NewUser(nil, username, string(digest), email)
	if err != nil {
		return fmt.Errorf("service<Account.Register>: %w", err)
	}

	if err := as.accountStore.Create(user); err != nil {
		return fmt.Errorf("service<Account.Register>: %w", err)
	}
	return nil
}

// =========================
// Event handling
// ========================

func (as Account) HandleRegisteredUsers(
	maxCount uint,
	serialize func(msg []messaging.UserRegistered) ([]byte, error),
) error {
	outbox, err := as.accountStore.GetRegisterOutbox(maxCount)
	if err != nil {
		return fmt.Errorf("service<Account.HandleNewUsers>: %w", err)
	} else if len(outbox) == 0 {
		return nil
	}

	payload, err := serialize(outbox)
	if err != nil {
		return fmt.Errorf("service<Account.HandleNewUsers>: %w", err)
	}

	err = as.messenger.Publish("default", payload, utility.NewDefaultAmqpPublishOpts(
		"", CreateSubscriptionQueue, "application/json"))
	if err != nil {
		return fmt.Errorf("service<Account.HandleNewUsers>: %w", err)
	}

	resolved := []uint64{}
	for _, o := range outbox {
		resolved = append(resolved, o.Id())
	}
	if err := as.accountStore.ResolveRegisterOutbox(resolved); err != nil {
		return fmt.Errorf("service<Account.HandleNewUsers>: %w", err)
	}
	return nil
}
