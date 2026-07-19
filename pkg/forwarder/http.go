package forwarder

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// TunnelMessage is one HTTP request serialized by the Tifybe edge and
// streamed down the WebSocket connection.
type TunnelMessage struct {
	Method  string            `json:"method"`
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

// Result describes what happened when a tunneled request was replayed
// against the local server.
type Result struct {
	Method   string
	Status   int
	Duration time.Duration
}

// Shared client: connection reuse matters when webhooks arrive in bursts.
var client = &http.Client{Timeout: 10 * time.Second}

// Forward replays a serialized tunnel message against targetURL and reports
// the local server's response status and how long it took.
func Forward(targetURL string, msg *TunnelMessage) (*Result, error) {
	req, err := http.NewRequest(msg.Method, targetURL, bytes.NewReader(msg.Body))
	if err != nil {
		return nil, err
	}

	for k, v := range msg.Headers {
		// The Host header must reflect the local target, not the tunnel edge.
		if http.CanonicalHeaderKey(k) == "Host" {
			continue
		}
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Drain so the connection can be reused.
	_, _ = io.Copy(io.Discard, resp.Body)

	return &Result{Method: msg.Method, Status: resp.StatusCode, Duration: time.Since(start)}, nil
}

// Parse decodes a raw WebSocket frame into a TunnelMessage. A frame that
// carries an "error" field (and no method) is a control message from the
// backend, returned via the second value.
func Parse(raw []byte) (*TunnelMessage, string, error) {
	var envelope struct {
		TunnelMessage
		Error string `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, "", err
	}
	if envelope.Error != "" && envelope.Method == "" {
		return nil, envelope.Error, nil
	}
	return &envelope.TunnelMessage, "", nil
}
