package messaging

const ShortConfiguredName = "short.configured"

type ShortConfigured struct {
	id          uint64
	userId      uint64
	linkId      uint64
	shortened   string
	destination string
}

func (sc ShortConfigured) Id() uint64          { return sc.id }
func (sc ShortConfigured) LinkId() uint64      { return sc.linkId }
func (sc ShortConfigured) UserId() uint64      { return sc.userId }
func (sc ShortConfigured) Shortened() string   { return sc.shortened }
func (sc ShortConfigured) Destination() string { return sc.destination }

func NewShortConfigured(
	id uint64,
	linkId uint64,
	userId uint64,
	shortened string,
	destination string,
) ShortConfigured {
	return ShortConfigured{
		id:          id,
		userId:      userId,
		linkId:      linkId,
		shortened:   shortened,
		destination: destination}
}
