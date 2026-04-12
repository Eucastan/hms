package security

import (
	"errors"
	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

const MinPasswordScore = 3

func ValidatePasswordStrength(password string) error {
	strength := zxcvbn.PasswordStrength(password, nil)
	if strength.Score < MinPasswordScore {
		return errors.New("password too weak")
	}

	return nil
}

func Hashpwd(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 12)
}

func Checkpwd(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}

	return true
}
