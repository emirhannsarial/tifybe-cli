package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

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

func init() {
	listenCmd.Flags().StringVar(&backendURLFlag, "backend-url", "wss://api.tifybe.com", "Tifybe backend websocket URL")
	listenCmd.Flags().StringVar(&frontendURLFlag, "frontend-url", "https://tifybe.com", "Tifybe frontend URL")
	listenCmd.Flags().StringVar(&subdomainFlag, "subdomain", "", "Use a persistent custom subdomain (requires authentication)")
	rootCmd.AddCommand(listenCmd)
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "req_" + hex.EncodeToString(b)
}

var listenCmd = &cobra.Command{
	Use:   "listen [port]",
	Short: "Listen for webhooks and forward to localhost",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		forwardID := ""
		if subdomainFlag != "" {
			if cfg.APIKey == "" {
				return fmt.Errorf("authentication required to use a custom subdomain. Please run `tifybe login` first")
			}
			if strings.HasPrefix(subdomainFlag, "req_") {
				return fmt.Errorf("custom subdomains cannot start with 'req_'")
			}
			forwardID = subdomainFlag
		} else {
			forwardID = generateID()
		}

		backendURL := fmt.Sprintf("%s/v1/cli/connect/%s", backendURLFlag, forwardID)

		apiBase := backendURLFlag
		if strings.HasPrefix(backendURLFlag, "wss://") {
			apiBase = "https://" + strings.TrimPrefix(backendURLFlag, "wss://")
		} else if strings.HasPrefix(backendURLFlag, "ws://") {
			apiBase = "http://" + strings.TrimPrefix(backendURLFlag, "ws://")
		}

		publicForwardURL := fmt.Sprintf("%s/local/%s", apiBase, forwardID)
		webViewerURL := fmt.Sprintf("%s/local/%s", frontendURLFlag, forwardID)

		fmt.Println("Tifybe CLI v1.0.0")
		fmt.Println("────────────────────────────────────────────────────────")
		fmt.Printf("Forwarding:      %s  ->  http://localhost:%s\n", publicForwardURL, port)
		fmt.Printf("Web Interface:   %s\n", webViewerURL)
		fmt.Println("Status:          🟢 Online & Listening")
		if subdomainFlag == "" && cfg.APIKey == "" {
			fmt.Println("\n💡 Tip: Want a persistent URL that never changes?")
			fmt.Println("   Run `tifybe login` to connect your free account.")
		}
		fmt.Println("\nWaiting for webhooks...")
		fmt.Println("────────────────────────────────────────────────────────")

		headers := make(http.Header)
		if cfg.APIKey != "" {
			headers.Set("Authorization", "Bearer "+cfg.APIKey)
		}

		c, resp, err := websocket.DefaultDialer.Dial(backendURL, headers)
		if err != nil {
			if resp != nil && resp.StatusCode == 403 {
				return fmt.Errorf("failed to connect: access denied. The subdomain might be claimed by another user or your API key is invalid")
			}
			return fmt.Errorf("failed to connect to Tifybe backend: %w", err)
		}
		defer c.Close()

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		done := make(chan struct{})
		targetURL := fmt.Sprintf("http://localhost:%s", port)

		go func() {
			defer close(done)
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("Connection closed:", err)
					return
				}

				// The backend might send an error message instead of closing immediately sometimes
				if strings.Contains(string(message), `"error":`) {
					log.Printf("Backend Error: %s", string(message))
					return
				}

				err = forwarder.Forward(targetURL, message)
				if err != nil {
					log.Printf("❌ Failed to forward: %v", err)
				} else {
					log.Printf("✅ Forwarded webhook to %s", targetURL)
				}
			}
		}()

		select {
		case <-interrupt:
			fmt.Println("\nShutting down due to interrupt...")
		case <-done:
			return fmt.Errorf("websocket connection dropped")
		}

		return nil
	},
}
