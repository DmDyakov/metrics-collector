package encryptor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptor_EncryptDecrypt(t *testing.T) {
	e, err := New("testdata/public.pem", "testdata/private.pem")
	if err != nil {
		t.Skip("test keys not found, skipping")
	}

	original := []byte("secret metrics data")
	encrypted, err := e.Encrypt(original)
	require.NoError(t, err)
	assert.NotEqual(t, original, encrypted)

	decrypted, err := e.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, original, decrypted)
}

func TestEncryptor_EncryptWithoutPublicKey(t *testing.T) {
	e := &Encryptor{}
	_, err := e.Encrypt([]byte("test"))
	assert.ErrorIs(t, err, ErrKeyNotSet)
}

func TestEncryptor_DecryptWithoutPrivateKey(t *testing.T) {
	e := &Encryptor{}
	_, err := e.Decrypt([]byte("test"))
	assert.ErrorIs(t, err, ErrKeyNotSet)
}

func TestNew_InvalidKeyPath(t *testing.T) {
	_, err := New("nonexistent.pem", "")
	assert.Error(t, err)
}
