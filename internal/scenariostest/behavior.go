package scenariostest

import (
	"net/http"
	"testing"
	"time"

	"github.com/Azure/httpx/internal/scenarios"

	"github.com/stretchr/testify/assert"
)

// ExpectServerToRun creates test scenarios for validating a server serving behaviors.
func ExpectServerToRun(
	t *testing.T,
	name string,
	CreateServer scenarios.CreateServerFunc,
	ApplyServer scenarios.ApplyServerFunc,
	ListenAndServe scenarios.ListenAndServeFunc,
) {
	t.Helper()

	const expectedBody = "Hello, world!"
	const expectedStatusCode = http.StatusOK

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedStatusCode)
		_, _ = w.Write([]byte(expectedBody))
	})

	expectServeByHandler := func(tc *TestContext, resp *http.Response, body []byte) {
		assert.Equal(tc, expectedStatusCode, resp.StatusCode, "status code")
		assert.Equal(tc, expectedBody, string(body), "response body")
	}

	t.Run(name+"/CreateServer", func(t *testing.T) {
		t.Parallel()

		tc := New(t)
		lis := tc.ListenTCP()
		addr := lis.Addr().String()
		s, err := CreateServer(addr, handler)
		assert.NoError(t, err, "CreateServerWith failed")
		assert.NotNil(t, s, "CreateServerWith returned nil")

		ctx, cancel := tc.CreateFastCheckContext()
		defer cancel()

		go func() {
			defer s.Close()
			_ = s.Serve(lis)
		}()

		tc.ExpectToServeResponse(ctx, addr, expectServeByHandler)

		shutdownErr := s.Shutdown(ctx)
		assert.NoError(t, shutdownErr, "Shutdown failed")
	})

	t.Run(name+"/ApplyServer", func(t *testing.T) {
		t.Parallel()

		tc := New(t)
		lis := tc.ListenTCP()
		addr := lis.Addr().String()

		s, err := ApplyServer(&http.Server{
			Addr:        addr,
			Handler:     handler,
			ReadTimeout: 0 * time.Second, // NOTE: workaround for G112/114
		})
		assert.NoError(t, err, "ApplyServerWith failed")
		assert.NotNil(t, s, "ApplyServerWith returned nil")

		ctx, cancel := tc.CreateFastCheckContext()
		defer cancel()

		go func() {
			defer s.Close()
			_ = s.Serve(lis)
		}()

		tc.ExpectToServeResponse(ctx, addr, expectServeByHandler)

		shutdownErr := s.Shutdown(ctx)
		assert.NoError(t, shutdownErr, "Shutdown failed")
	})

	t.Run(name+"/ListenAndServe", func(t *testing.T) {
		t.Parallel()

		go func() {
			err := ListenAndServe("127.0.0.1:0", handler)
			assert.NoError(t, err, "ListenAndStartServerWith failed")
		}()

		// ListenAndServe test case we wait for the server to start without issue after 500ms
		time.Sleep(500 * time.Millisecond)
	})
}
