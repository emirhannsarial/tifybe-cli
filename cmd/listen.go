package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"github.com/emirhannsarial/tifybe-cli/pkg/config"
	"github.com/emirhannsarial/tifybe-cli/pkg/forwarder"
)

var (
	backendURLFlag  string
	frontendURLFlag string
	subdomainFlag   string
)

const (
	reconnectMin = 1 * time.Second
	reconnectMax = 30 * time.Second

	// pongWait is how long the edge may go silent before the session is
	// treated as dead. It has to exceed pingPeriod by enough that one lost
	// ping does not end a healthy session.
	pongWait = 70 * time.Second

	// pingPeriod is how often we prove we are still here.
	pingPeriod = 30 * time.Second

	// writeWait bounds a single control-frame write so a wedged socket cannot
	// block the pinger.
	writeWait = 10 * time.Second
)

func init() {
	listenCmd.Flags().StringVar(&backendURLFlag, "backend-url", "wss://api.tifybe.com", "Tifybe backend WebSocket URL")
	listenCmd.Flags().StringVar(&frontendURLFlag, "frontend-url", "https://tifybe.com", "Tifybe web inspector URL")
	listenCmd.Flags().StringVar(&subdomainFlag, "subdomain", "", "claim a persistent subdomain (requires `tifybe login`)")
	rootCmd.AddCommand(listenCmd)
}

func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("could not generate session id: %w", err)
	}
	return "req_" + hex.EncodeToString(b), nil
}

func stamp() string {
	return time.Now().Format("15:04:05")
}

var listenCmd = &cobra.Command{
	Use:     "listen <port>",
	Short:   "Listen for webhooks and forward them to localhost",
	Example: "  tifybe listen 8080\n  tifybe listen 3000 --subdomain=my-startup",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port := args[0]
		if _, err := strconv.Atoi(port); err != nil {
			return fmt.Errorf("port must be a number, got %q", port)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var forwardID string
		if subdomainFlag != "" {
			if cfg.APIKey == "" {
				return fmt.Errorf("custom subdomains require authentication — run `tifybe login` first")
			}
			if strings.HasPrefix(subdomainFlag, "req_") {
				return fmt.Errorf("custom subdomains cannot start with %q", "req_")
			}
			forwardID = subdomainFlag
		} else {
			forwardID, err = generateID()
			if err != nil {
				return err
			}
		}

		wsURL := fmt.Sprintf("%s/v1/cli/connect/%s", backendURLFlag, forwardID)

		apiBase := backendURLFlag
		if strings.HasPrefix(backendURLFlag, "wss://") {
			apiBase = "https://" + strings.TrimPrefix(backendURLFlag, "wss://")
		} else if strings.HasPrefix(backendURLFlag, "ws://") {
			apiBase = "http://" + strings.TrimPrefix(backendURLFlag, "ws://")
		}

		publicURL := fmt.Sprintf("%s/local/%s", apiBase, forwardID)
		inspectorURL := fmt.Sprintf("%s/local/%s", frontendURLFlag, forwardID)
		targetURL := "http://localhost:" + port

		fmt.Printf("tifybe %s\n\n", Version)
		fmt.Printf("  Forwarding   %s -> %s\n", publicURL, targetURL)
		fmt.Printf("  Inspector    %s\n", inspectorURL)
		if subdomainFlag == "" && cfg.APIKey == "" {
			fmt.Printf("\n  tip: run `tifybe login` to claim a persistent URL that survives restarts\n")
		}
		fmt.Println()

		headers := make(http.Header)
		if cfg.APIKey != "" {
			headers.Set("Authorization", "Bearer "+cfg.APIKey)
		}

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		// Reconnect loop: a dropped connection retries with exponential
		// backoff instead of killing the session. Auth failures are fatal.
		delay := reconnectMin
		for {
			err := runSession(wsURL, targetURL, headers, interrupt)
			switch {
			case err == nil: // interrupted by the user
				fmt.Printf("%s  session closed\n", stamp())
				return nil
			case isFatal(err):
				return err
			}

			fmt.Printf("%s  connection lost (%v) — reconnecting in %s\n", stamp(), err, delay)
			select {
			case <-interrupt:
				return nil
			case <-time.After(delay):
			}
			if delay *= 2; delay > reconnectMax {
				delay = reconnectMax
			}
		}
	},
}

type fatalError struct{ error }

func isFatal(err error) bool {
	_, ok := err.(fatalError)
	return ok
}

// runSession dials the tunnel and forwards messages until the connection
// drops (returned as an error) or the user interrupts (returned as nil).
func runSession(wsURL, targetURL string, headers http.Header, interrupt <-chan os.Signal) error {
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		if resp != nil && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized) {
			return fatalError{fmt.Errorf("access denied: the subdomain may be claimed by another user, or your API key is invalid")}
		}
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	fmt.Printf("%s  connected — waiting for webhooks\n", stamp())

	// Keepalive. Two jobs, and the second is the one that matters.
	//
	// Proxies drop idle connections, so we ping. But a connection can also die
	// without ever closing — the laptop suspends, the wifi changes, a NAT entry
	// expires. Reads then block on it forever while writes keep succeeding into
	// the kernel buffer, so `tifybe listen` sits there printing nothing and
	// looking perfectly healthy while every webhook goes into a socket that no
	// longer leads anywhere.
	//
	// The read deadline is what tells the difference. Every frame the edge
	// sends — pong, ping or a webhook — pushes it forward; silence past
	// pongWait ends the session, and the reconnect loop takes it from there.
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return fmt.Errorf("could not arm read deadline: %w", err)
	}
	extendDeadline := func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	}
	conn.SetPongHandler(extendDeadline)
	conn.SetPingHandler(func(appData string) error {
		_ = extendDeadline(appData)
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(writeWait))
	})

	stopPing := make(chan struct{})
	defer close(stopPing)
	go func() {
		t := time.NewTicker(pingPeriod)
		defer t.Stop()
		for {
			select {
			case <-stopPing:
				return
			case <-t.C:
				// A protocol ping the edge answers automatically, so the
				// deadline advances without the backend needing to change.
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
					return
				}
				// The text heartbeat older edges reply "pong" to.
				if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
					return
				}
			}
		}
	}()

	done := make(chan error, 1)
	go func() {
		for {
			_, raw, err := conn.ReadMessage()
			// Any frame is evidence the edge is still there, webhooks
			// included — not just the keepalive replies.
			_ = extendDeadline("")
			if err != nil {
				done <- err
				return
			}
			if string(raw) == "pong" {
				continue
			}

			msg, ctrlErr, err := forwarder.Parse(raw)
			if err != nil {
				fmt.Printf("%s  skipped an unparseable frame: %v\n", stamp(), err)
				continue
			}
			if ctrlErr != "" {
				done <- fatalError{fmt.Errorf("backend: %s", ctrlErr)}
				return
			}

			res, err := forwarder.Forward(targetURL, msg)
			if err != nil {
				fmt.Printf("%s  %s -> %s  FAILED: %v\n", stamp(), msg.Method, targetURL, err)
				continue
			}
			fmt.Printf("%s  %s -> %s  %d (%dms)\n",
				stamp(), res.Method, targetURL, res.Status, res.Duration.Milliseconds())
		}
	}()

	select {
	case <-interrupt:
		// Best-effort graceful close so the edge frees the subdomain quickly.
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(2*time.Second))
		return nil
	case err := <-done:
		return err
	}
}
