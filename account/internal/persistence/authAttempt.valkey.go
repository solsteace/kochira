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

type authAttemptRow struct {
	Ok   bool
	Time time.Time
}

func (row authAttemptRow) useTime() int64 {
	return row.Time.UnixMilli()
}

func (row authAttemptRow) toHash() map[string]string {
	return map[string]string{
		"ok":   fmt.Sprintf("%t", row.Ok),
		"time": fmt.Sprintf("%d", row.useTime())}
}

func (_ authAttemptRow) fromHash(h map[string]string) (authAttemptRow, error) {
	hOk, err := strconv.ParseBool(h["ok"])
	if err != nil {
		return authAttemptRow{}, fmt.Errorf("persistence<authAttemptRow.fromHash>: %w", err)
	}
	hTime, err := strconv.ParseInt(h["time"], 10, 64)
	if err != nil {
		return authAttemptRow{}, fmt.Errorf("persistence<authAttemptRow.fromHash>: %w", err)
	}

	row := authAttemptRow{
		Ok:   hOk,
		Time: time.UnixMilli(hTime)}
	return row, nil
}

func (row authAttemptRow) toDomain() (auth.Attempt, error) {
	a, err := auth.NewAttempt(row.Ok, row.Time)
	if err != nil {
		return auth.Attempt{}, fmt.Errorf("persistence<authAttemptRow.toDomain>: %w", err)
	}
	return a, nil
}

func newAuthAttemptRow(a auth.Attempt) authAttemptRow {
	return authAttemptRow{a.Ok(), a.Time()}
}

type valkeyAuthAttempt struct {
	client        valkey.Client
	retentionTime time.Duration
}

func NewValkeyAuthAttempt(client valkey.Client, retentionTime time.Duration) valkeyAuthAttempt {
	return valkeyAuthAttempt{client, retentionTime}
}

func (vaa valkeyAuthAttempt) Add(userId uint, aa auth.Attempt) error {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(vaa.client)
	tx := adapter.TxPipeline()

	identifier := rand.Text()
	zName := fmt.Sprintf("user:%d:auth-attempts", userId)
	hName := fmt.Sprintf("user:%d:auth-attempts:%s", userId, identifier)

	row := newAuthAttemptRow(aa)
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
		return fmt.Errorf("persistence<valkeyAuthAttempt.Add>: %w", err)
	}
	return nil
}

func (vaa valkeyAuthAttempt) Get(userId uint) ([]auth.Attempt, error) {
	ctx := context.Background()
	adapter := valkeycompat.NewAdapter(vaa.client)

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

		row, err := authAttemptRow{}.fromHash(res.Val())
		if err != nil {
			return []auth.Attempt{}, fmt.Errorf("persistence<valkeyAuthAttempt.Get>: %w", err)
		}
		attempt, _ := row.toDomain()
		attempts = append(attempts, attempt)
	}
	return attempts, nil
}
