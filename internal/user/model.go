package user

import "golang.org/x/crypto/bcrypt"

// User is a registered person that can login to the
// application. Password is the encrypted password.
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type BcryptHash struct {
	value []byte
}

func NewBcryptHash(password string) (BcryptHash, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return BcryptHash{}, err
	}
	return BcryptHash{hash}, err
}
