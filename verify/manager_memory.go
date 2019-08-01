package verify

var _ Manager = new(ManagerMemory)

type ManagerMemory struct {
}

func NewManagerMemory() *ManagerMemory {
	return &ManagerMemory{}
}

func (m *ManagerMemory) Verify(key string) error {
	panic("")
}

func (m *ManagerMemory) UnVerify(key string) error {
	panic("")
}

func (m *ManagerMemory) IsVerified(key string) (bool, error) {
	panic("")
}
