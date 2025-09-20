package service

import (
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/account/internal/domain/auth"
)

type Jailer struct {
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

func NewJailer(
	maxConsecFail int,
	maxExtraFail int,
	baseJailTime time.Duration,
	extraJailTime time.Duration,
) Jailer {
	return Jailer{
		maxConsecFail,
		maxExtraFail,
		baseJailTime,
		extraJailTime}
}

// Returns recommended cache retention time in seconds. Basically answers "how
// long should I remember your failed attempt?"
func (j Jailer) RetentionTime(leeway time.Duration) time.Duration {
	if leeway < 0 {
		leeway = 0
	}
	return j.baseJailTime +
		time.Duration(j.maxExtraFail)*j.extraJailTime +
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
func (j Jailer) IsJailed(attempts []auth.Attempt) error {
	endIdx := len(attempts)
	if len(attempts) > (j.maxConsecFail + j.maxExtraFail) {
		endIdx = j.maxConsecFail + j.maxExtraFail
	}

	now := time.Now()
	allowedConsecFails := j.maxConsecFail
	var jailTime time.Duration
	for _, a := range attempts[:int(endIdx)] {
		if a.Ok() {
			break
		}
		allowedConsecFails--

		switch {
		case allowedConsecFails == 0:
			oldestDiff := a.Time().Sub(now) + j.baseJailTime
			jailTime = oldestDiff
		case allowedConsecFails < 0:
			jailTime += j.extraJailTime
		}
	}

	if jailTime > 0 {
		err := oops.Unauthorized{
			Msg: fmt.Sprintf(
				"Failed too many times! Try again in %.2fs", jailTime.Seconds())}
		return fmt.Errorf("service<AuthJailer.IsJailed> %w", err)
	}
	return nil
}
