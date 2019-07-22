package verify

// We need to be able to remove keys if they have been removed yeah smart thinking aeneas

type Manager interface {
	Verify(key string) error
	UnVerify(key string) error
	IsVerified(key string) (bool, error)
}
