package scenarios

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplyServerWith_override_existing(t *testing.T) {
	s := &http.Server{
		ReadTimeout: 10 * time.Second,
	}
	apply := ApplyServerWith(ConfigureServer(func(s *http.Server) (*http.Server, error) {
		s.ReadTimeout = 1 * time.Second
		s.ReadHeaderTimeout = 3 * time.Second

		return s, nil
	}))
	updatedServer, err := apply(s)
	assert.NoError(t, err)
	assert.Equal(t, 1*time.Second, updatedServer.ReadTimeout)
	assert.Equal(t, 3*time.Second, updatedServer.ReadHeaderTimeout)
}

func TestApplyServerWith_configure_error(t *testing.T) {
	s := &http.Server{
		ReadTimeout: 10 * time.Second,
	}
	apply := ApplyServerWith(ConfigureServer(func(s *http.Server) (*http.Server, error) {
		return s, assert.AnError
	}))
	updatedServer, err := apply(s)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, updatedServer)
}

func TestCreateServerWith_override_existing(t *testing.T) {
	create := CreateServerWith(ConfigureServer(func(s *http.Server) (*http.Server, error) {
		s.ReadTimeout = 1 * time.Second
		s.ReadHeaderTimeout = 3 * time.Second

		return s, nil
	}))
	updatedServer, err := create("127.0.0.1:0", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1*time.Second, updatedServer.ReadTimeout)
	assert.Equal(t, 3*time.Second, updatedServer.ReadHeaderTimeout)
	assert.NotEmpty(t, updatedServer.Addr)
}

func TestCreateServerWith_configure_error(t *testing.T) {
	create := CreateServerWith(ConfigureServer(func(s *http.Server) (*http.Server, error) {
		return s, assert.AnError
	}))
	updatedServer, err := create("127.0.0.1:0", nil)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, updatedServer)
}

func TestListenAndServeFuncWith_error(t *testing.T) {
	lis := ListenAndServeFuncWith(ConfigureServer(func(s *http.Server) (*http.Server, error) {
		return s, assert.AnError
	}))
	err := lis("127.0.0.1:0", nil)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestListenAndServeFuncWith_succeeded(t *testing.T) {
	lis := ListenAndServeFuncWith(ConfigureServer(func(s *http.Server) (*http.Server, error) {
		s.ReadHeaderTimeout = 10 * time.Second

		go func() {
			time.Sleep(500 * time.Millisecond)
			_ = s.Close()
		}()

		return s, nil
	}))
	_ = lis("127.0.0.1:0", nil)
}
