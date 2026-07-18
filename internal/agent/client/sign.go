package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"metrics-collector/internal/errs"
	"net/http"

	"go.uber.org/zap"
)

func (c *Client) createSignature(data []byte) []byte {
	hmacHash := hmac.New(sha256.New, []byte(c.secretKey))
	hmacHash.Write(data)
	return hmacHash.Sum(nil)

}

func (c *Client) checkResponseSignature(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body",
			zap.Error(err),
			zap.Int("body_size", len(body)))
		return &errs.SignatureError{Msg: "failed to read response body"}
	}

	received, err := hex.DecodeString(resp.Header.Get("HashSHA256"))
	if err != nil {
		return &errs.SignatureError{Msg: "invalid signature format"}
	}

	expected := c.createSignature(body)

	if !hmac.Equal(received, expected) {
		c.logger.Error("signed body mismatch",
			zap.String("expected", hex.EncodeToString(expected)),
			zap.String("received", hex.EncodeToString(received)),
		)
		return &errs.SignatureError{Msg: "invalid response signature"}
	}

	return nil
}
