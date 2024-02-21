package app

import (
	"net/http"
	"time"

	"github.com/Azure/httpx/internal/scenarios"
)

var serverSettings = []scenarios.ServerConfig{
	scenarios.ConfigureServer(func(s *http.Server) (*http.Server, error) {
		// set short read header timeout to protect against slowloris attacks
		s.ReadHeaderTimeout = 5 * time.Second
		// set this value to protect against malicious attacks. Use 60 seconds to match Nginx's default client_body_timeout:
		// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout
		s.ReadTimeout = 60 * time.Second

		// set this value to protect against malicious attacks. Use 60 seconds to match ARM client request timeout:
		// https://github.com/Azure/azure-resource-manager-rpc/blob/master/v1.0/common-api-details.md#client-request-timeout
		s.WriteTimeout = 60 * time.Second

		// TODO:
		// - consider wrapping incoming request with a read size limit: per ARM spec, the maximum size of a request is 4MB
		// - consider wrapping outgoing response with a write size limit: per ARM spec, the maximum size of a response is 8MB

		return s, nil
	}),
}

// CreateServer creates a http.Server instance for application server usage.
var CreateServer = scenarios.CreateServerWith(serverSettings...)

// ApplyServer applies configurations to a http.Server instance for application server usage.
var ApplyServer = scenarios.ApplyServerWith(serverSettings...)

// ListenAndServe listens on the TCP network address addr and then calls Serve to handle requests on incoming connections for application server usage.
var ListenAndServe = scenarios.ListenAndServeFuncWith(serverSettings...)
