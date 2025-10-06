package persistence

import (
	"context"
	"crypto/rand"
	"fmt"
	"strconv"
	"time"

	"github.com/solsteace/kochira/account/internal/domain/auth"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"
)

type valkeyAuth struct {
	client           valkey.Client
	attemptRetention time.Duration
	tokenRetention   time.Duration
}

func NewValkeyAuth(
	client valkey.Client,
	attemptRetention time.Duration,
	tokenRetention time.Duration,
) valkeyAuth {
	return valkeyAuth{
		client:           client,
		attemptRetention: attemptRetention,
		tokenRetention:   tokenRetention}
}

// =====================================
//  Token
// =====================================

func (va valkeyAuth) Add(userId uint, aa auth.Attempt) error {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(va.client)
	tx := adapter.TxPipeline()

	identifier := rand.Text()
	zName := fmt.Sprintf("user:%d:auth-attempts", userId)
	hName := fmt.Sprintf("user:%d:auth-attempts:%s", userId, identifier)

	row := newAuthAttempt(aa)
	zData := valkeycompat.Z{
		Member: hName,
		Score:  float64(row.Time.UnixMilli())}
	hData := row.toHash()

	// Most of the time, we only interested with the latest data, so the Z being
	// expired together with the data pointed by its latest member should be good enough
	tx.ZAdd(ctx, zName, zData)
	tx.Expire(ctx, zName, va.attemptRetention)
	tx.HSet(ctx, hName, hData)
	tx.Expire(ctx, hName, va.attemptRetention)
	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("persistence<valkeyAuthAttempt.Add>: %w", err)
	}
	return nil
}

func (va valkeyAuth) Get(userId uint) ([]auth.Attempt, error) {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(va.client)

	zName := fmt.Sprintf("user:%d:auth-attempts", userId)
	attemptKeys := adapter.ZRevRange(ctx, zName, 0, -1)
	if attemptKeys.Err() != nil {
		return []auth.Attempt{}, fmt.Errorf("persistence<valkeyAuthAttempt.Get>: %w", attemptKeys.Err())
	}

	attempts := []auth.Attempt{}
	for _, attemptKey := range attemptKeys.Val() {
		res := adapter.HGetAll(ctx, attemptKey)
		if len(res.Val()) == 0 {
			continue
		}

		row, err := valkeyAuthAttempt{}.fromHash(res.Val())
		if err != nil {
			return []auth.Attempt{}, fmt.Errorf("persistence<valkeyAuthAttempt.Get>: %w", err)
		}
		attempt, _ := row.toDomain()
		attempts = append(attempts, attempt)
	}
	return attempts, nil
}

// =====================================
//  AUTH ATTEMPT
// =====================================

type valkeyAuthAttempt struct {
	Ok   bool
	Time time.Time
}

func newAuthAttempt(a auth.Attempt) valkeyAuthAttempt {
	return valkeyAuthAttempt{a.Ok(), a.Time()}
}

func (row valkeyAuthAttempt) toHash() map[string]string {
	return map[string]string{
		"ok":   fmt.Sprintf("%t", row.Ok),
		"time": fmt.Sprintf("%d", row.Time.UnixMilli())}
}

func (_ valkeyAuthAttempt) fromHash(hash map[string]string) (valkeyAuthAttempt, error) {
	hashOk, err := strconv.ParseBool(hash["ok"])
	if err != nil {
		return valkeyAuthAttempt{}, fmt.Errorf("persistence<authAttemptRow.fromHash>: %w", err)
	}
	hashTime, err := strconv.ParseInt(hash["time"], 10, 64)
	if err != nil {
		return valkeyAuthAttempt{}, fmt.Errorf("persistence<authAttemptRow.fromHash>: %w", err)
	}

	row := valkeyAuthAttempt{
		Ok:   hashOk,
		Time: time.UnixMilli(hashTime)}
	return row, nil
}

func (row valkeyAuthAttempt) toDomain() (auth.Attempt, error) {
	a, err := auth.NewAttempt(row.Ok, row.Time)
	if err != nil {
		return auth.Attempt{}, fmt.Errorf("persistence<authAttemptRow.toDomain>: %w", err)
	}
	return a, nil
}

func (va valkeyAuth) Grant(userId uint, token string) error {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(va.client)

	key := fmt.Sprintf("user:%d:auth-token", userId)
	res := adapter.Set(ctx, key, token, va.tokenRetention)
	if res.Err() != nil {
		return fmt.Errorf("persistence<valkeyToken.Grant>: %w", res.Err())
	}
	return nil
}

func (va valkeyAuth) Revoke(userId uint) error {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(va.client)

	key := fmt.Sprintf("user:%d:auth-token", userId)
	res := adapter.Del(ctx, key)
	if res.Err() != nil {
		return fmt.Errorf("persistence<valkeyToken.Revoke>: %w", res.Err())
	}
	return nil
}

func (va valkeyAuth) FindByOwner(userId uint) (string, error) {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(va.client)

	key := fmt.Sprintf("user:%d:auth-token", userId)
	res := adapter.Get(ctx, key)
	if res.Err() != nil {
		return "", fmt.Errorf("persistence<valkeyToken.FindByOwner>: %w", res.Err())
	}
	return res.String(), nil
}
