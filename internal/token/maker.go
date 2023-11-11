package token

import "time"

// Maker is the interface that wraps the basic Token methods.
type Maker interface {
	CreateToken(username string, role string, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
