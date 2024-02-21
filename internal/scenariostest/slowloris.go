package scenariostest

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Azure/httpx/internal/scenarios"

	"github.com/b4fun/turtle"
	"github.com/stretchr/testify/require"
)

func ValidateSlowloris(
	t *testing.T,
	name string,
	CreateServer scenarios.CreateServerFunc,
	expectToPass bool,
) {
	t.Helper()

	t.Run(name+"/slowloris", func(t *testing.T) {
		// slowloris needs more than 1 minute to complete...
		skipForFastEnvironment(t)

		t.Parallel()

		tc := New(t)
		lis := tc.ListenTCP()
		addr := lis.Addr().String()
		s, err := CreateServer(addr, http.NotFoundHandler())
		require.NoError(t, err, "CreateServerWith failed")

		var connStateCounter = new(ConnStateCounter)
		s.ConnState = func(conn net.Conn, state http.ConnState) {
			// NOTE: we will see new -> active -> closed state transition
			// ref: https://cs.opensource.google/go/go/+/refs/tags/go1.20.7:src/net/http/server.go;l=1846;drc=ea4631cc0cf301c824bd665a7980c13289ab5c9d
			connStateCounter.Count(conn, state)
		}

		// FIXME: consider make this change to be configurable?
		// We are expecting the provided server is configured with ReadHeaderTimeout <= 60s
		duration := 90 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		go func() {
			defer s.Close()
			_ = s.Serve(lis)
		}()

		targetURL, err := url.Parse("http://" + addr)
		require.NoError(tc, err, "parse url failed")

		client := turtle.Slowloris{
			Target: turtle.Target{
				Url:      *targetURL,
				Duration: duration,
				// it's intended to set 1 connection, we only verify the server
				// will close the connection for slow header. For load test, please use turtle cli / slowhttptest.
				Connections: 1,
			},
			SendGibberish:     true,
			GibberishInterval: 100 * time.Millisecond,
			WriteTimeout:      1 * time.Second,
		}

		go func() {
			slowlorisErr := client.Run(ctx)
			if slowlorisErr != nil {
				// for debugging purpose only
				log.Printf("slowloris failed: %s", slowlorisErr)
			}
		}()

		<-ctx.Done()

		// if the server is slowloris proof, we expect to see:
		// - at least one full state transition
		// - at least two active states
		countNewStateLoaded := connStateCounter.CountStateNew.Load()
		countActiveStateLoaded := connStateCounter.CountActiveState.Load()
		countClosedStateLoaded := connStateCounter.CountClosedState.Load()
		isSlowlorisProof := countNewStateLoaded > 1 && countActiveStateLoaded > 0 && countClosedStateLoaded > 0
		tc.Logf(
			"countNewState: %d, countActiveState: %d, countClosedState: %d",
			countNewStateLoaded, countActiveStateLoaded, countClosedStateLoaded,
		)
		if expectToPass {
			require.True(tc, isSlowlorisProof, "server should be slowloris proof")
		} else {
			require.False(tc, isSlowlorisProof, "server should not be slowloris proof")
		}
	})
}
