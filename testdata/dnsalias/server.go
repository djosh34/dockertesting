// Package dnsalias provides a simple HTTP server for testing DNS alias resolution.
package dnsalias

import (
	"fmt"
	"net/http"
)

// DNSAlias is the expected DNS alias that the container must be configured with.
// When running in dockertest, the container should have this alias set via WithAliases().
const DNSAlias = "myapp.test"

// DefaultPort is the default port the test server listens on.
const DefaultPort = 8080

// HelloHandler returns a handler that responds with a greeting message.
func HelloHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello from %s!", DNSAlias)
	}
}

// StartServer starts an HTTP server on the given port and returns a shutdown function.
// The server listens on all interfaces (0.0.0.0) so it's accessible via the DNS alias.
func StartServer(port int) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HelloHandler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return server, nil
}
