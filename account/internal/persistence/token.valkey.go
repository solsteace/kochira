package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"
)

type valkeyToken struct {
	client        valkey.Client
	tokenLifetime time.Duration
}

func NewValkeyToken(
	client valkey.Client,
	tokenLifetime time.Duration,
) valkeyToken {
	return valkeyToken{client, tokenLifetime}
}

func (vt valkeyToken) Grant(userId uint, token string) error {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(vt.client)

	key := fmt.Sprintf("user:%d:auth-token", userId)
	res := adapter.Set(ctx, key, token, vt.tokenLifetime)
	if res.Err() != nil {
		return fmt.Errorf("persistence<valkeyToken.Grant>: %w", res.Err())
	}
	return nil
}

func (vt valkeyToken) Revoke(userId uint) error {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(vt.client)

	key := fmt.Sprintf("user:%d:auth-token", userId)
	res := adapter.Del(ctx, key)
	if res.Err() != nil {
		return fmt.Errorf("persistence<valkeyToken.Revoke>: %w", res.Err())
	}
	return nil
}

func (vt valkeyToken) FindByOwner(userId uint) (string, error) {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(vt.client)

	key := fmt.Sprintf("user:%d:auth-token", userId)
	res := adapter.Get(ctx, key)
	if res.Err() != nil {
		return "", fmt.Errorf("persistence<valkeyToken.FindByOwner>: %w", res.Err())
	}
	return res.String(), nil
}
