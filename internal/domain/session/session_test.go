package session

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/nbf-auth/internal/domain/user"
	"github.com/stretchr/testify/require"
)

func TestValidateAccessToken(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	uid := uuid.New()
	user, err := user.NewUser(uid, "123", "John", "Doe")
	require.NoError(t, err)

	accessTTL := 10 * time.Minute
	refreshTTL := 10 * time.Minute
	tokens, err := GenerateTokens(user, privateKey, accessTTL, refreshTTL)
	require.NoError(t, err)

	newUser, err := tokens.ValidateAccessToken(&privateKey.PublicKey)
	require.NoError(t, err)

	require.Equal(t, user.ID, newUser.ID)
	require.Equal(t, user.Name, newUser.Name)
	require.Equal(t, user.Surname, newUser.Surname)
}
