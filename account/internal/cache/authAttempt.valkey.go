package cache

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/solsteace/kochira/account/internal/domain"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"
)

type valkeyAuthAttempt struct {
	client        valkey.Client
	retentionTime time.Duration
}

func NewValkeyAuthAttempt(
	client valkey.Client,
	retentionTime time.Duration,
) valkeyAuthAttempt {
	return valkeyAuthAttempt{client, retentionTime}
}

func (vaa valkeyAuthAttempt) Add(userId uint, aa domain.AuthAttempt) error {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(vaa.client)
	tx := adapter.TxPipeline()

	identifier := rand.Text()
	zName := fmt.Sprintf("user:%d:auth-attempts", userId)
	hName := fmt.Sprintf("user:%d:auth-attempts:%s", userId, identifier)

	row := NewAuthAttemptRow(aa)
	zData := valkeycompat.Z{
		Member: hName,
		Score:  float64(row.useTime())}
	hData := row.toHash()

	// Most of the time, we only interested with the latest data, so the Z being
	// expired together with the data pointed by its latest member should be good enough
	tx.ZAdd(ctx, zName, zData)
	tx.Expire(ctx, zName, vaa.retentionTime)
	tx.HSet(ctx, hName, hData)
	tx.Expire(ctx, hName, vaa.retentionTime)
	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("cache<valkeyAuthAttempt.Add>: %w", err)
	}
	return nil
}

func (vaa valkeyAuthAttempt) Get(userId uint) ([]domain.AuthAttempt, error) {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(vaa.client)

	zName := fmt.Sprintf("user:%d:auth-attempts", userId)
	attemptKeys := adapter.ZRevRange(ctx, zName, 0, -1)
	if attemptKeys.Err() != nil {
		return []domain.AuthAttempt{}, fmt.Errorf("cache<valkeyAuthAttempt.Get>: %w", attemptKeys.Err())
	}

	attempts := []domain.AuthAttempt{}
	for _, attemptKey := range attemptKeys.Val() {
		res := adapter.HGetAll(ctx, attemptKey)
		if len(res.Val()) == 0 {
			continue
		}

		row, err := authAttemptRow{}.fromHash(res.Val())
		if err != nil {
			return []domain.AuthAttempt{}, fmt.Errorf("cache<valkeyAuthAttempt.Get>: %w", err)
		}
		attempt, _ := row.toDomain()
		attempts = append(attempts, attempt)
	}
	return attempts, nil
}
