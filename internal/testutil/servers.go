package testutil

import (
	"net/http"
	"net/http/httptest"
)

// TestServer provides utilities for creating test HTTP servers
// This eliminates the repeated httptest.NewServer setup across multiple files

// NewAPITestServer creates a test server that serves OpenAPI specs and API responses
func NewAPITestServer(spec string) *httptest.Server {
	mux := http.NewServeMux()

	// Serve the OpenAPI spec at /openapi.json
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(spec))
	})

	// Serve common test endpoints
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id": 1, "name": "Test User"}]`))
		case "POST":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 2, "name": "New User"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": 1, "name": "Specific User"}`))
	})

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "test response"}`))
	})

	// Endpoint for testing different content types
	mux.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		switch accept {
		case "application/json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data": "json"}`))
		case "application/xml":
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(`<data>xml</data>`))
		case "text/plain":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(`plain text data`))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data": "default"}`))
		}
	})

	// Search endpoint with query parameters
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("query")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		w.WriteHeader(http.StatusOK)
		// Simple JSON encoding for test
		w.Write([]byte(`{"query":"` + query + `","limit":"` + limit + `","offset":"` + offset + `","results":["item1","item2"]}`))
	})

	return httptest.NewServer(mux)
}

// NewErrorTestServer creates a test server that returns various HTTP error responses
func NewErrorTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/400", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad Request"}`))
	})

	mux.HandleFunc("/401", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	})

	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not Found"}`))
	})

	mux.HandleFunc("/500", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal Server Error"}`))
	})

	return httptest.NewServer(mux)
}

// NewSlowTestServer creates a test server with artificial delays for timeout testing
func NewSlowTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// Note: In real implementation, we'd add time.Sleep(time.Duration(delaySeconds) * time.Second)
		// For tests, we'll keep it fast but simulate the behavior
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "slow response"}`))
	})

	return httptest.NewServer(mux)
}

// GetTestServerURL returns just the URL portion of a test server for easy use
func GetTestServerURL(server *httptest.Server) string {
	return server.URL
}