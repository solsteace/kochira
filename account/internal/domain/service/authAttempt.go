package service

import (
	"time"

	"github.com/solsteace/kochira/account/internal/domain"
)

type AuthAttempt struct {
	// Number of consecutive fail attempts. If reached, jail time would be activated
	maxConsecFail int

	// Number of failed attempts to consider beyond `MaxConsecFail`
	maxExtraFail int

	// The jail time in seconds added when failed attempt count had reached
	// `MaxConsecFail` for the first time
	baseJailTime time.Duration

	// The jail time in seconds added for every extra failed attempt
	extraJailTime time.Duration
}

func NewAuthAttempt(
	maxConsecFail int,
	maxExtraFail int,
	baseJailTime time.Duration,
	extraJailTime time.Duration,
) AuthAttempt {
	return AuthAttempt{
		maxConsecFail,
		maxExtraFail,
		baseJailTime,
		extraJailTime}
}

// Returns recommended cache retention time in seconds. Basically answers "how
// long should I remember your failed attempt?"
func (aa AuthAttempt) RetentionTime(leeway time.Duration) time.Duration {
	if leeway < 0 {
		leeway = 0
	}
	return aa.baseJailTime +
		time.Duration(aa.maxExtraFail)*aa.extraJailTime +
		leeway
}

// Calculates the jail time for authentication. It works like this:
//
// 1. An attempt is considered failed if `ok` is false
//
// 2. If the consecutive failed attempt count had reached `aa.maxExtraFail`, add
// `aa.BaseJailTime` counted since the oldest failed attempt within checked attempts.
// Otherwise, return 0
//
// 3. If there're still more failed attempts, add `aa.ExtraJailTime` until a non-failed
// attempt had been found or reached the end of checked attempts
func (aa AuthAttempt) CalculateJailTime(attempts []domain.AuthAttempt) time.Duration {
	endIdx := len(attempts)
	if len(attempts) > (aa.maxConsecFail + aa.maxExtraFail) {
		endIdx = aa.maxConsecFail + aa.maxExtraFail
	}

	now := time.Now()
	allowedConsecFails := aa.maxConsecFail
	var jailTime time.Duration
	for _, a := range attempts[:int(endIdx)] {
		if a.Ok() {
			break
		}
		allowedConsecFails--

		switch {
		case allowedConsecFails == 0:
			oldestDiff := a.Time().Sub(now) + aa.baseJailTime
			jailTime = oldestDiff
		case allowedConsecFails < 0:
			jailTime += aa.extraJailTime
		}
	}
	return jailTime
}
