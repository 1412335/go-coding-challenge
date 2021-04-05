package error

import "github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"

var (
	ErrMissingEmail   = errors.BadRequest("Email is required", map[string]string{"email": "Missing email"})
	ErrInvalidEmail   = errors.BadRequest("Invalid email", map[string]string{"email": "The email provided is invalid"})
	ErrDuplicateEmail = errors.BadRequest("Duplicate email", map[string]string{"email": "A user with this email address already exists"})

	ErrInvalidPassword   = errors.BadRequest("Invalid password", map[string]string{"password": "Password must be at least 8 characters long"})
	ErrIncorrectPassword = errors.Unauthenticated("Email or password is incorrect", "password", "Email or password is incorrect")
	ErrHashPassword      = errors.InternalServerError("Hash password failed", "hash password failed")

	ErrMissingUserID    = errors.BadRequest("Missing user id", map[string]string{"id": "Missing user id"})
	ErrMissingAccountID = errors.BadRequest("Missing account id", map[string]string{"id": "Missing account id"})

	ErrInvalidTransactionAmount = errors.BadRequest("Invalid transaction amount", map[string]string{"amount": "greater than zero"})

	ErrConnectDB = errors.InternalServerError("Connect db failed", "Connecting to database failed")

	ErrUserNotFound = errors.NotFound("Not found user", map[string]string{"user": "User not found"})

	ErrMissingToken   = errors.BadRequest("Token missing", map[string]string{"token": "Missing token"})
	ErrTokenGenerated = errors.InternalServerError("Token gen failed", "Generate token failed")
	ErrTokenInvalid   = errors.Unauthenticated("Invalid token", "token", "Token invalid")
)
