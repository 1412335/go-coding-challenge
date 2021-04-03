package user

import "regexp"

var (
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func isValidEmail(email string) bool {
	if len(email) == 0 || len(email) > 255 {
		return false
	}
	return emailRegex.MatchString(email)
}

func isValidPassword(password string) bool {
	return len(password) >= 8
}
