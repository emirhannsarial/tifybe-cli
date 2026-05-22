package forwarder

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type TunnelMessage struct {
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

func Forward(targetURL string, rawMessage []byte) error {
	var msg TunnelMessage
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		return err
	}

	req, err := http.NewRequest(msg.Method, targetURL, bytes.NewReader(msg.Body))
	if err != nil {
		return err
	}

	for k, v := range msg.Headers {
		// Avoid overriding host header to prevent issues with localhost servers
		if k != "Host" {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
