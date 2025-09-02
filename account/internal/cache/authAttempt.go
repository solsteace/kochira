package cache

import (
	"fmt"
	"strconv"
	"time"

	"github.com/solsteace/kochira/account/internal/domain"
)

type AuthAttempt interface {
	Add(userId uint, row domain.AuthAttempt) error
	Get(userId uint) ([]domain.AuthAttempt, error)
}

type authAttemptRow struct {
	Ok   bool
	Time time.Time
}

func NewAuthAttemptRow(aa domain.AuthAttempt) authAttemptRow {
	return authAttemptRow{aa.Ok(), aa.Time()}
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
		fmt.Println(err)
		return authAttemptRow{}, err
	}
	hTime, err := strconv.ParseInt(h["time"], 10, 64)
	if err != nil {
		return authAttemptRow{}, err
	}

	row := authAttemptRow{
		Ok:   hOk,
		Time: time.UnixMilli(hTime)}
	return row, nil
}

func (row authAttemptRow) toDomain() (domain.AuthAttempt, error) {
	aa, _ := domain.NewAuthAttempt(row.Ok, row.Time)
	return aa, nil
}
