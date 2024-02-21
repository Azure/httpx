package scenarios

import "net/http"

// CreateServerFunc creates a http.Server instance.
type CreateServerFunc func(addr string, handler http.Handler) (*http.Server, error)

// CreateServerWith creates a CreateServerFunc with the given ServerConfig.
func CreateServerWith(
	configures ...ServerConfig,
) CreateServerFunc {
	apply := ApplyServerWith(configures...)

	return func(addr string, handler http.Handler) (*http.Server, error) {
		server := new(http.Server)
		var err error
		server, err = apply(server)
		if err != nil {
			return nil, err
		}

		server.Addr = addr
		server.Handler = handler

		return server, nil
	}
}

// ApplyServerFunc applies configurations to a http.Server instance.
type ApplyServerFunc func(*http.Server) (*http.Server, error)

// ApplyServerWith creates an ApplyServerFunc with the given ServerConfig.
func ApplyServerWith(
	configures ...ServerConfig,
) ApplyServerFunc {
	return func(s *http.Server) (*http.Server, error) {
		for _, configure := range configures {
			var err error
			s, err = configure.configureServer(s)
			if err != nil {
				return nil, err
			}
		}

		return s, nil
	}
}

// ListenAndServeFunc listens on the TCP network address addr and then calls Serve to handle requests on incoming connections.
type ListenAndServeFunc func(addr string, handler http.Handler) error

// ListenAndServeFunc creates a ListenAndServeFunc with the given ServerConfig.
func ListenAndServeFuncWith(
	configures ...ServerConfig,
) ListenAndServeFunc {
	create := CreateServerWith(configures...)
	return func(addr string, handler http.Handler) error {
		server, err := create(addr, handler)
		if err != nil {
			return err
		}

		return server.ListenAndServe()
	}
}
