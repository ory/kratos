package hash

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
)

// Upgrader provides methods for upgrading password hash algorithm.
type Upgrader struct {
	c UpgraderConfiguration
}

type UpgraderConfiguration interface {
	config.Provider
}

func NewHashUpgrader(c UpgraderConfiguration) *Upgrader {
	return &Upgrader{c: c}
}

// DoesNeedToUpgrade returns given hashed password needs to upgrade new password hash algorithm.
func (u *Upgrader) DoesNeedToUpgrade(ctx context.Context, hash []byte) bool {
	_, ok := u.getUpgradeHasherByHash(ctx, hash)
	return ok
}

// Upgrade returns a hash derived from the password by new password hash algorithm.
func (u *Upgrader) Upgrade(ctx context.Context, hash []byte, password []byte) ([]byte, error) {
	hasher, ok := u.getUpgradeHasherByHash(ctx, hash)
	if !ok {
		return nil, errors.Errorf("Not found new hasher method.")
	}
	return hasher.Generate(ctx, password)
}

func (u *Upgrader) getUpgradeHasherByHash(ctx context.Context, hash []byte) (Hasher, bool) {
	if IsPbkdf2Hash(hash) {
		return u.getHasherByAlgorithm(u.c.Config(ctx).HasherPbkdf2().UpgradeTo)
	}
	return nil, false
}

func (u *Upgrader) getHasherByAlgorithm(alg string) (Hasher, bool) {
	switch alg {
	case "bcrypt":
		return NewHasherBcrypt(u.c), true
	case "argon2":
		return NewHasherArgon2(u.c), true
	default:
		return nil, false
	}
}
