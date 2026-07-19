package forwarder_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emirhannsarial/tifybe-cli/pkg/forwarder"
)

func TestForward(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Header.Get("X-Test") != "Value" {
			t.Errorf("expected header X-Test: Value")
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	msg := &forwarder.TunnelMessage{
		Method:  "POST",
		Headers: map[string]string{"X-Test": "Value"},
		Body:    []byte(`{"test":true}`),
	}

	res, err := forwarder.Forward(srv.URL, msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected server to be called")
	}
	if res.Status != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", res.Status)
	}
}

func TestParse(t *testing.T) {
	raw, _ := json.Marshal(forwarder.TunnelMessage{Method: "POST", Body: []byte(`{}`)})
	msg, ctrlErr, err := forwarder.Parse(raw)
	if err != nil || ctrlErr != "" {
		t.Fatalf("unexpected: err=%v ctrl=%q", err, ctrlErr)
	}
	if msg.Method != "POST" {
		t.Errorf("expected POST, got %q", msg.Method)
	}
}

func TestParseControlError(t *testing.T) {
	_, ctrlErr, err := forwarder.Parse([]byte(`{"error":"subdomain taken"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctrlErr != "subdomain taken" {
		t.Errorf("expected control error, got %q", ctrlErr)
	}
}

// A webhook whose *payload* contains an "error" key must be forwarded, not
// treated as a backend control message (regression: the old implementation
// killed the tunnel on such payloads).
func TestParsePayloadContainingErrorKey(t *testing.T) {
	raw, _ := json.Marshal(forwarder.TunnelMessage{
		Method: "POST",
		Body:   []byte(`{"error":"card_declined"}`),
	})
	msg, ctrlErr, err := forwarder.Parse(raw)
	if err != nil || ctrlErr != "" {
		t.Fatalf("payload with error key misclassified: err=%v ctrl=%q", err, ctrlErr)
	}
	if msg == nil || msg.Method != "POST" {
		t.Fatalf("expected a normal tunnel message")
	}
}
