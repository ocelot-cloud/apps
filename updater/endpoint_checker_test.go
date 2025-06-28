package main

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTryAccessingIndexPageOnLocalhost(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	port := s.URL[len("http://localhost:"):]
	ec := EndpointCheckerImpl{}
	err := ec.TryAccessingIndexPageOnLocalhost(port, "/")
	require.NoError(t, err)
}
