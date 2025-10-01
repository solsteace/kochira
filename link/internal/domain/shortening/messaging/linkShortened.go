package messaging

const LinkShortenedName = "link.shortened"

type LinkShortened struct {
	id     uint64
	userId uint64
	linkId uint64
}

func (ls LinkShortened) Id() uint64     { return ls.id }
func (ls LinkShortened) UserId() uint64 { return ls.userId }
func (ls LinkShortened) LinkId() uint64 { return ls.linkId }

func NewLinkShortened(id, userId, linkId uint64) LinkShortened {
	return LinkShortened{
		id:     id,
		userId: userId,
		linkId: linkId}
}
