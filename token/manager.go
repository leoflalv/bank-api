package token

import "time"

// Interface for managing tokens
type Manager interface {
	// Creates a new token for a specific username and duration
	CreateToken(username string, duration time.Duration) (string, error)

	// Check if the token is valid
	VerifyToken(token string) (*Payload, error)
}
