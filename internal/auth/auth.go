package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword takes a plain-text password and returns its bcrypt hash.
func HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword handles salting and hashing.
	// The second argument is the "cost", which determines how much
	// computational effort is used. A higher cost is more secure
	// but slower. bcrypt.DefaultCost is a good starting point.
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

// CheckPasswordHash compares a plain-text password with a stored bcrypt hash.
// It returns nil if they match, and an error if they don't.
func CheckPasswordHash(password, hash string) error {
	// This function is specifically designed to be "timing-attack resistant",
	// meaning it takes a constant amount of time to run, regardless of
	// whether the password is correct or not. This prevents attackers from
	// gaining information by measuring response times.
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
