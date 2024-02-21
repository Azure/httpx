package scenariostest

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Azure/httpx/internal/scenarios"

	"github.com/b4fun/turtle"
	"github.com/stretchr/testify/require"
)

func ValidateSlowBodyRead(
	t *testing.T,
	name string,
	CreateServer scenarios.CreateServerFunc,
	expectToPass bool,
) {
	t.Helper()

	t.Run(name+"/slow-body-read", func(t *testing.T) {
		skipForFastEnvironment(t)

		t.Parallel()

		tc := New(t)
		lis := tc.ListenTCP()
		addr := lis.Addr().String()
		var countTimeoutErr atomic.Int64
		s, err := CreateServer(addr, http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			require.NotNil(tc, req.Body, "incoming request should provide body")
			defer func() {
				_ = req.Body.Close()
			}()

			wr.WriteHeader(http.StatusOK)

			_, readBodyErr := io.Copy(io.Discard, req.Body)
			if readBodyErr != nil {
				log.Printf("read body failed: %s", readBodyErr)
				if os.IsTimeout(readBodyErr) {
					countTimeoutErr.Add(1)
				}
			}
		}))
		require.NoError(t, err, "CreateServerWith failed")

		var connStateCounter = new(ConnStateCounter)
		s.ConnState = func(conn net.Conn, state http.ConnState) {
			// NOTE: we will see new -> active -> closed state transition
			// ref: https://cs.opensource.google/go/go/+/refs/tags/go1.20.7:src/net/http/server.go;l=1846;drc=ea4631cc0cf301c824bd665a7980c13289ab5c9d
			connStateCounter.Count(conn, state)
		}

		// FIXME: consider make this change to be configurable?
		// We are expecting the provided server is configured with ReadBodyTimeout <= 60s
		duration := 90 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		go func() {
			defer s.Close()
			_ = s.Serve(lis)
		}()

		targetURL, err := url.Parse("http://" + addr)
		require.NoError(tc, err, "parse url failed")

		client := turtle.SlowBodyReadRequest{
			Target: turtle.Target{
				Url:         *targetURL,
				Duration:    duration,
				Connections: 1,
			},
			Method: http.MethodPost,
		}

		go func() {
			runErr := client.Run(ctx)
			if runErr != nil {
				// for debugging purpose only
				log.Printf("slow body read failed: %s", runErr)
			}
		}()

		<-ctx.Done()

		// if the server is slow body read proof, we expect to see:
		// - at least one full state transition
		// - at least two active states
		// - at least one timeout error
		countNewStateLoaded := connStateCounter.CountStateNew.Load()
		countActiveStateLoaded := connStateCounter.CountActiveState.Load()
		countClosedStateLoaded := connStateCounter.CountClosedState.Load()
		countTimeoutErrLoaded := countTimeoutErr.Load()
		isSlowBodyReadProof := countNewStateLoaded > 1 && countActiveStateLoaded > 0 && countClosedStateLoaded > 0 && countTimeoutErrLoaded > 0
		tc.Logf(
			"countNewState: %d, countActiveState: %d, countClosedState: %d, countTimeoutErr: %d",
			countNewStateLoaded, countActiveStateLoaded, countClosedStateLoaded, countTimeoutErrLoaded,
		)
		if expectToPass {
			require.True(tc, isSlowBodyReadProof, "server should be slow body read proof")
		} else {
			require.False(tc, isSlowBodyReadProof, "server should not be slow body read proof")
		}
	})

	t.Run(name+"/slow-body-write", func(t *testing.T) {
		// TODO(hbc): figure out how to test this settings
		t.Skip("slow body write involves buffer settings between the client and the server, which is not easy to test at this moment")
	})
}
