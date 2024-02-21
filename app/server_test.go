package app

import (
	"testing"

	"github.com/Azure/httpx/internal/scenariostest"
)

func TestServer_Behavior(t *testing.T) {
	const scenarioName = "app"

	scenariostest.ExpectServerToRun(t, scenarioName, CreateServer, ApplyServer, ListenAndServe)

	// countNewState: 18, countActiveState: 17, countClosedState: 17
	scenariostest.ValidateSlowloris(t, scenarioName, CreateServer, true)

	// countNewState: 2, countActiveState: 2, countClosedState: 1
	// The body read time out is 60s, we run the test for 90s.
	// In this case, the server should accept 2 connections and close at least 1 for read body time out
	// before the test is finished.
	scenariostest.ValidateSlowBodyRead(t, scenarioName, CreateServer, true)
}
