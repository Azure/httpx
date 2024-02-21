package scenariostest_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Azure/httpx/internal/scenariostest"
)

func Test_ValidateSlow_Vulnerable(t *testing.T) {
	createVulnerableServer := func(addr string, handler http.Handler) (*http.Server, error) {
		server := &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		}

		return server, nil
	}

	// countNewState: 1, countActiveState: 0, countClosedState: 0
	scenariostest.ValidateSlowBodyRead(t, "vulnerable", createVulnerableServer, false)
}
