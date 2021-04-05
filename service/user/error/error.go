package error

import "github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"

var (
	ErrMissingEmail   = errors.BadRequest("MISSING_EMAIL", map[string]string{"email": "Missing email"})
	ErrInvalidEmail   = errors.BadRequest("INVALID_EMAIL", map[string]string{"email": "The email provided is invalid"})
	ErrDuplicateEmail = errors.BadRequest("DUPLICATE_EMAIL", map[string]string{"email": "A user with this email address already exists"})

	ErrInvalidPassword   = errors.BadRequest("INVALID_PASSWORD", map[string]string{"password": "Password must be at least 8 characters long"})
	ErrIncorrectPassword = errors.Unauthenticated("INCORRECT_PASSWORD", "password", "Email or password is incorrect")
	ErrHashPassword      = errors.InternalServerError("HASH_PASSWORD", "hash password failed")

	ErrMissingUserID    = errors.BadRequest("MISSING_ID", map[string]string{"id": "Missing user id"})
	ErrMissingAccountID = errors.BadRequest("MISSING_ACCOUNT_ID", map[string]string{"id": "Missing account id"})

	ErrInvalidTransactionAmount = errors.BadRequest("INVALID_TRANSACTION", map[string]string{"amount": "greater than zero"})

	ErrConnectDB = errors.InternalServerError("CONNECT_DB", "Connecting to database failed")

	ErrUserNotFound = errors.NotFound("NOT_FOUND", map[string]string{"user": "User not found"})

	ErrMissingToken   = errors.BadRequest("MISSING_TOKEN", map[string]string{"token": "Missing token"})
	ErrTokenGenerated = errors.InternalServerError("TOKEN_GEN_FAILED", "Generate token failed")
	ErrTokenInvalid   = errors.Unauthenticated("TOKEN_INVALID", "token", "Token invalid")
)
