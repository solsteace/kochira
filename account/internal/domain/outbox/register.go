package outbox

type Register struct {
	id     uint64
	userId uint64
	isDone bool
}

func (ro *Register) Done() { ro.isDone = true }

func (ro Register) Id() uint64     { return ro.id }
func (ro Register) UserId() uint64 { return ro.userId }
func NewRegister(id, userId uint64, isDone bool) Register {
	return Register{id, userId, isDone}
}
