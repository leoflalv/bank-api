package token

import (
	"github/leoflalv/bank-api/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPaseto(t *testing.T) {
	manager, err := NewPasetoManager(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomString(10)
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := manager.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := manager.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.Id)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredPaseto(t *testing.T) {
	manager, err := NewPasetoManager(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomString(10)

	token, err := manager.CreateToken(username, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := manager.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}
