package publisher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testEvent struct {
	Message string `json:"message"`
}

func TestNewPublisher(t *testing.T) {
	t.Run("file publisher", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "audit-*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		pub, err := NewPublisher[testEvent](tmpFile.Name(), "", time.Second, zap.NewNop())
		require.NoError(t, err)
		assert.NotNil(t, pub)
	})

	t.Run("http publisher", func(t *testing.T) {
		pub, err := NewPublisher[testEvent]("", "http://localhost:8080", time.Second, zap.NewNop())
		require.NoError(t, err)
		assert.NotNil(t, pub)
	})

	t.Run("both disabled returns nil", func(t *testing.T) {
		pub, err := NewPublisher[testEvent]("", "", time.Second, zap.NewNop())
		assert.NoError(t, err)
		assert.Nil(t, pub)
	})

	t.Run("invalid file path", func(t *testing.T) {
		pub, err := NewPublisher[testEvent]("/invalid/path/file.log", "", time.Second, zap.NewNop())
		assert.Error(t, err)
		assert.Nil(t, pub)
	})
}

func TestPublisher_Run(t *testing.T) {
	t.Run("nil publisher", func(t *testing.T) {
		var pub *Publisher[testEvent]
		err := pub.Run(context.Background())
		assert.NoError(t, err)
	})

	t.Run("graceful shutdown", func(t *testing.T) {
		pub, err := NewPublisher[testEvent]("", "http://localhost:8080", time.Second, zap.NewNop())
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err = pub.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("publishes to file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "audit-*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		pub, err := NewPublisher[testEvent](tmpFile.Name(), "", time.Second, zap.NewNop())
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			_ = pub.Run(ctx)
		}()

		event := testEvent{Message: "hello"}
		pub.Publish(event)

		time.Sleep(100 * time.Millisecond)
		cancel()
		time.Sleep(100 * time.Millisecond)

		content, err := os.ReadFile(tmpFile.Name())
		require.NoError(t, err)
		assert.NotEmpty(t, content)
	})

	t.Run("publishes to http", func(t *testing.T) {
		received := make(chan testEvent, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var event testEvent
			if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
				t.Logf("decode error: %v", err)
			}
			received <- event
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		pub, err := NewPublisher[testEvent]("", server.URL, time.Second, zap.NewNop())
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			_ = pub.Run(ctx)
		}()

		event := testEvent{Message: "http test"}
		pub.Publish(event)

		select {
		case receivedEvent := <-received:
			assert.Equal(t, event, receivedEvent)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for event")
		}

		cancel()
	})
}

func TestPublisher_Publish(t *testing.T) {
	t.Run("publish after stop does not panic", func(t *testing.T) {
		pub, err := NewPublisher[testEvent]("", "http://localhost:8080", time.Second, zap.NewNop())
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			_ = pub.Run(ctx)
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()
		time.Sleep(50 * time.Millisecond)

		assert.NotPanics(t, func() {
			pub.Publish(testEvent{Message: "after stop"})
		})
	})
}
