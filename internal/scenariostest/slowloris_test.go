package scenariostest_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Azure/httpx/internal/scenariostest"
)

func Test_ValidateSlowloris_Vulnerable(t *testing.T) {
	createVulnerableServer := func(addr string, handler http.Handler) (*http.Server, error) {
		server := &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 1 * time.Second,
		}
		server.ReadHeaderTimeout = 0 // hack to bypass lint check

		return server, nil
	}

	// countNewState: 1, countActiveState: 0, countClosedState: 0
	scenariostest.ValidateSlowloris(t, "vulnerable", createVulnerableServer, false)
}
