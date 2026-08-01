// Package encryptor предоставляет шифрование и расшифровку данных с использованием RSA.
package encryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// Encryptor шифрует и расшифровывает данные с помощью RSA-ключей.
type Encryptor struct {
	publicKey  *rsa.PublicKey  // для шифрования
	privateKey *rsa.PrivateKey // для расшифровки
}

// ErrEncryptionFailed возникает при ошибке шифрования данных.
var ErrEncryptionFailed = errors.New("encryption failed")

// ErrDecryptionFailed возникает при ошибке расшифровки данных.
var ErrDecryptionFailed = errors.New("decryption failed")

// ErrKeyNotSet возникает, если ключ не был загружен.
var ErrKeyNotSet = errors.New("key not set")

// New создаёт Encryptor. Если передан путь к публичному ключу - используется для шифрования.
// Если передан путь к приватному ключу - используется для расшифровки.
// Можно передать оба пути.
func New(publicKeyPath, privateKeyPath string) (*Encryptor, error) {
	if publicKeyPath == "" && privateKeyPath == "" {
		return nil, nil
	}

	e := &Encryptor{}

	if publicKeyPath != "" {
		pub, err := readPublicKey(publicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key: %w", err)
		}
		e.publicKey = pub
	}

	if privateKeyPath != "" {
		priv, err := readPrivateKey(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
		e.privateKey = priv
	}

	return e, nil
}

// Encrypt шифрует данные
func (e *Encryptor) Encrypt(data []byte) ([]byte, error) {
	if e.publicKey == nil {
		return nil, fmt.Errorf("%w: public key not set", ErrKeyNotSet)
	}

	maxSize := e.publicKey.Size() - 2*sha256.Size - 2
	if len(data) > maxSize {
		return nil, fmt.Errorf("%w: message too large (%d > %d bytes)",
			ErrEncryptionFailed,
			len(data),
			maxSize,
		)
	}

	encrypted, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		e.publicKey,
		data,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	return encrypted, nil
}

// Decrypt расшифровывает данные после Encrypt.
func (e *Encryptor) Decrypt(data []byte) ([]byte, error) {
	if e.privateKey == nil {
		return nil, fmt.Errorf("%w: private key not set", ErrKeyNotSet)
	}

	decrypted, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		e.privateKey,
		data,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return decrypted, nil
}

// readPublicKey читает публичный ключ из PEM-файла.
func readPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid public key PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

// readPrivateKey читает приватный ключ из PEM-файла.
func readPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key PEM")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
