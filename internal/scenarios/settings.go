package scenarios

import "net/http"

// ServerConfig configures a http.Server instance.
type ServerConfig interface {
	configureServer(s *http.Server) (*http.Server, error)
}

// ConfigureServer configures a http.Server instance.
type ConfigureServer func(s *http.Server) (*http.Server, error)

var _ ServerConfig = ConfigureServer(nil)

func (f ConfigureServer) configureServer(s *http.Server) (*http.Server, error) {
	return f(s)
}
