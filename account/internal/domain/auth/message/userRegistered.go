package message

type UserRegistered struct {
	id     uint64
	userId uint64
	isDone bool
}

func (ur *UserRegistered) Done() { ur.isDone = true }

func (ur UserRegistered) Id() uint64     { return ur.id }
func (ur UserRegistered) UserId() uint64 { return ur.userId }

func NewRegister(id, userId uint64, isDone bool) UserRegistered {
	return UserRegistered{id, userId, isDone}
}
