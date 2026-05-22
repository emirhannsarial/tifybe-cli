package forwarder_test

import (
	"encoding/json"
	"github.com/tifybe/tifybe-cli/pkg/forwarder"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForward(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Header.Get("X-Test") != "Value" {
			t.Errorf("expected header X-Test: Value")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	msg := forwarder.TunnelMessage{
		Method:  "POST",
		Headers: map[string]string{"X-Test": "Value"},
		Body:    []byte(`{"test":true}`),
	}

	raw, _ := json.Marshal(msg)
	err := forwarder.Forward(srv.URL, raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Errorf("expected server to be called")
	}
}
