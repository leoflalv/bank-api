package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoManager struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoManager(symetricKey string) (Manager, error) {
	if len(symetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: key size should be exactly %d characters", chacha20poly1305.KeySize)
	}

	manager := &PasetoManager{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symetricKey),
	}

	return manager, nil
}

func (manager *PasetoManager) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	return manager.paseto.Encrypt(manager.symmetricKey, payload, nil)
}

// Check if the token is valid
func (manager *PasetoManager) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := manager.paseto.Decrypt(token, manager.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
