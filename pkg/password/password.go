package password

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(plainText string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainText), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

// CheckPassword compares a plain-text password against a bcrypt hashed password.
func CheckPassword(hashed, plainText string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plainText))
}
