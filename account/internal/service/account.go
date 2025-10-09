package service

import (
	"fmt"

	"github.com/solsteace/kochira/account/internal/domain/account"
	accountStore "github.com/solsteace/kochira/account/internal/domain/account/store"
	"github.com/solsteace/kochira/account/internal/utility/hash"
)

const CreateSubscriptionQueue = "subscription.creator" // `depends on `subscription` service

type Account struct {
	accountStore accountStore.User
	hasher       hash.Handler
}

func NewAccount(
	store accountStore.User,
	hasher hash.Handler,
) Account {
	return Account{
		accountStore: store,
		hasher:       hasher}
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
	payloader func(users []uint64) ([]byte, error),
	send func([]byte) error,
) error {
	outbox, err := as.accountStore.GetRegisterOutbox(maxCount)
	if err != nil {
		return fmt.Errorf("service<Account.HandleNewUsers>: %w", err)
	} else if len(outbox) == 0 {
		return nil
	}

	handledUser := []uint64{}
	for _, o := range outbox {
		handledUser = append(handledUser, o.UserId())
	}
	payload, err := payloader(handledUser)
	if err != nil {
		return fmt.Errorf("service<Account.HandleNewUsers>: %w", err)
	}

	if err := send(payload); err != nil {
		return fmt.Errorf("service<Account.HandleNewUsers>: %w", err)
	}

	if err := as.accountStore.ResolveRegisterOutbox(handledUser); err != nil {
		return fmt.Errorf("service<Account.HandleNewUsers>: %w", err)
	}
	return nil
}
