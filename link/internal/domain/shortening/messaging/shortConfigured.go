package messaging

const ShortConfiguredName = "short.configured"

type ShortConfigured struct {
	id          uint64
	userId      uint64
	linkId      uint64
	alias       string
	destination string
	isOpen      bool
}

func (sc ShortConfigured) Id() uint64          { return sc.id }
func (sc ShortConfigured) LinkId() uint64      { return sc.linkId }
func (sc ShortConfigured) UserId() uint64      { return sc.userId }
func (sc ShortConfigured) Alias() string       { return sc.alias }
func (sc ShortConfigured) Destination() string { return sc.destination }
func (sc ShortConfigured) IsOpen() bool        { return sc.isOpen }

func NewShortConfigured(
	id uint64,
	linkId uint64,
	userId uint64,
	shortened string,
	destination string,
	isOpen bool,
) ShortConfigured {
	return ShortConfigured{
		id:          id,
		userId:      userId,
		linkId:      linkId,
		alias:       shortened,
		destination: destination,
		isOpen:      isOpen}
}
