package scenariostest

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Environment variable to skip slow tests.
// This is useful for CI/CD pipelines to skip slow tests.
const HTTPX_FAST = "HTTPX_FAST"

func isFastEnvironment(t testing.TB) bool {
	if testing.Short() {
		return true
	}

	if _, exists := os.LookupEnv(HTTPX_FAST); exists {
		return true
	}

	return false
}

func skipForFastEnvironment(t testing.TB) {
	t.Helper()

	if isFastEnvironment(t) {
		t.Skip("skipped slow test")
	}
}

type TestContext struct {
	testing.TB
}

// New creates a TestContext instance.
func New(t testing.TB) *TestContext {
	return &TestContext{TB: t}
}

// ListenTCP listens on the local TCP network with random port and returns the address.
// The connection will be closed when the test is finished.
func (tc *TestContext) ListenTCP() net.Listener {
	tc.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(tc, err, "ListenTCP failed")

	tc.Cleanup(func() {
		_ = lis.Close()
	})

	addr := lis.Addr().String()
	tc.Logf("ListenTCP: %q", addr)

	return lis
}

func (tc *TestContext) CreateFastCheckContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (tc *TestContext) ExpectToServeResponse(
	ctx context.Context,
	addr string,
	checkResp func(tc *TestContext, resp *http.Response, body []byte),
) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr, nil)
	require.NoError(tc, err, "NewRequestWithContext failed")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(tc, err, "send request")

	var body []byte
	if resp.Body != nil {
		var readErr error
		body, readErr = io.ReadAll(resp.Body)
		require.NoError(tc, readErr, "read response body")
		require.NoError(tc, resp.Body.Close(), "close response body")
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	checkResp(tc, resp, body)
}

type ConnStateCounter struct {
	CountStateNew    atomic.Int64
	CountActiveState atomic.Int64
	CountClosedState atomic.Int64
}

func (c *ConnStateCounter) Count(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		c.CountStateNew.Add(1)
	case http.StateActive:
		c.CountActiveState.Add(1)
	case http.StateClosed:
		c.CountClosedState.Add(1)
	}
}
