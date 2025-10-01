package value

import "time"

type Perk struct {
	lifetime       time.Duration // How long would a link last since its first opening?
	limit          uint          // How many simultaneously-active-links are allowed?
	allowShortEdit bool          // Does the shortened URL allowed to be customized?
}

func (p Perk) Lifetime() time.Duration { return p.lifetime }
func (p Perk) Limit() uint             { return p.limit }
func (p Perk) AllowShortEdit() bool    { return p.allowShortEdit }

func NewPerks(
	lifetime time.Duration,
	limit uint,
	allowShortEdit bool,
) Perk {
	return Perk{lifetime, limit, allowShortEdit}
}
