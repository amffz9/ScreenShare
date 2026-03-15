package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	qrcode "github.com/skip2/go-qrcode"
)

//go:embed public/*
var publicFS embed.FS

const (
	pingInterval = 30 * time.Second
	pongWait     = 35 * time.Second
)

// Config holds server configuration.
type Config struct {
	Port       int    `json:"port"`
	STUNServer string `json:"stunServer"`
}

// Client represents a connected WebSocket peer.
type Client struct {
	ID   string
	Conn *websocket.Conn
	Type string // "broadcaster" or "viewer"
	Name string
	mu   sync.Mutex
}

// BroadcasterInfo is the public info sent to viewers.
type BroadcasterInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IncomingMessage is the envelope for all client messages.
// Offer/Answer/Candidate use json.RawMessage to relay verbatim.
type IncomingMessage struct {
	Type      string          `json:"type"`
	Name      string          `json:"name,omitempty"`
	To        string          `json:"to,omitempty"`
	Offer     json.RawMessage `json:"offer,omitempty"`
	Answer    json.RawMessage `json:"answer,omitempty"`
	Candidate json.RawMessage `json:"candidate,omitempty"`
}

var (
	clients      = make(map[string]*Client)
	broadcasters = make(map[string]BroadcasterInfo)
	mu           sync.RWMutex
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	// --- Config: defaults -> config file -> CLI flags ---
	cfg := Config{
		Port:       8080,
		STUNServer: "stun:stun.l.google.com:19302",
	}

	configPath := flag.String("config", "config.json", "path to config file")
	flagPort := flag.Int("port", 0, "server port (overrides config file)")
	flagSTUN := flag.String("stun", "", "STUN server (overrides config file)")
	noOpen := flag.Bool("no-open", false, "don't open browser on startup")
	flag.Parse()

	// Load config file if it exists
	if data, err := os.ReadFile(*configPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Printf("Warning: could not parse %s: %v", *configPath, err)
		} else {
			log.Printf("Loaded config from %s", *configPath)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("Warning: could not read %s: %v", *configPath, err)
	} else {
		log.Printf("No config file found at %s, using defaults (port=%d, stun=%s)", *configPath, cfg.Port, cfg.STUNServer)
	}

	// CLI flags override
	if *flagPort != 0 {
		cfg.Port = *flagPort
	}
	if *flagSTUN != "" {
		cfg.STUNServer = *flagSTUN
	}

	// --- Routes ---
	publicContent, _ := fs.Sub(publicFS, "public")
	fileServer := http.FileServer(http.FS(publicContent))

	mux := http.NewServeMux()

	// /ip endpoint
	mux.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ip": getLocalIP()})
	})

	// /config endpoint — exposes client-relevant settings
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"stunServer": cfg.STUNServer})
	})

	// /qr endpoint — generates a QR code PNG pointing to the viewer page
	mux.HandleFunc("/qr", func(w http.ResponseWriter, r *http.Request) {
		viewerURL := fmt.Sprintf("http://%s:%d/viewer", getLocalIP(), cfg.Port)
		png, err := qrcode.Encode(viewerURL, qrcode.Medium, 256)
		if err != nil {
			http.Error(w, "failed to generate QR code", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(png)
	})

	// Route rewrites for clean URLs
	mux.HandleFunc("/broadcaster", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/broadcaster.html"
		fileServer.ServeHTTP(w, r)
	})
	mux.HandleFunc("/viewer", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/viewer.html"
		fileServer.ServeHTTP(w, r)
	})

	// Root: WebSocket upgrade or static file serving
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			handleWebSocket(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	// --- Server ---
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		ip := getLocalIP()
		log.Printf("Server running at http://localhost:%d", cfg.Port)
		log.Printf("Local Network Access: http://%s:%d", ip, cfg.Port)
		log.Printf("  Broadcaster: /broadcaster")
		log.Printf("  Viewer:      /viewer")

		if !*noOpen {
			openBrowser(fmt.Sprintf("http://localhost:%d", cfg.Port))
		}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")

	// Close all WebSocket connections
	mu.RLock()
	for _, c := range clients {
		c.Conn.Close()
	}
	mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

// handleWebSocket upgrades the connection and runs the read loop.
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	clientID := generateID()
	client := &Client{ID: clientID, Conn: conn}

	log.Printf("New connection: %s", clientID)

	// Configure pong handler to reset read deadline
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start ping ticker
	pingTicker := time.NewTicker(pingInterval)
	pingDone := make(chan struct{})
	defer func() {
		pingTicker.Stop()
		close(pingDone)
	}()

	go func() {
		for {
			select {
			case <-pingTicker.C:
				client.mu.Lock()
				err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
				client.mu.Unlock()
				if err != nil {
					return
				}
			case <-pingDone:
				return
			}
		}
	}()

	defer func() {
		conn.Close()
		mu.Lock()
		c, exists := clients[clientID]
		delete(clients, clientID)
		if exists && c.Type == "broadcaster" {
			delete(broadcasters, clientID)
			log.Printf("Broadcaster removed: %s", c.Name)
			mu.Unlock()
			broadcastToViewers(map[string]interface{}{
				"type": "broadcaster-left",
				"id":   clientID,
			})
		} else {
			mu.Unlock()
		}
		log.Printf("Client disconnected: %s", clientID)
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg IncomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		switch msg.Type {
		case "register-broadcaster":
			name := msg.Name
			if name == "" {
				name = "Unnamed Broadcaster"
			}
			client.Type = "broadcaster"
			client.Name = name

			mu.Lock()
			clients[clientID] = client
			broadcasters[clientID] = BroadcasterInfo{ID: clientID, Name: name}
			mu.Unlock()

			log.Printf("Broadcaster registered: %s (%s)", name, clientID)
			sendJSON(client, map[string]interface{}{
				"type": "registered",
				"id":   clientID,
				"name": name,
			})
			broadcastToViewers(map[string]interface{}{
				"type":        "broadcaster-joined",
				"broadcaster": BroadcasterInfo{ID: clientID, Name: name},
			})

		case "register-viewer":
			client.Type = "viewer"
			mu.Lock()
			clients[clientID] = client
			mu.Unlock()

			log.Printf("Viewer registered: %s", clientID)
			sendJSON(client, map[string]interface{}{
				"type": "registered",
				"id":   clientID,
			})

		case "get-broadcasters":
			if client.Type == "" {
				log.Printf("Ignoring get-broadcasters from unregistered client %s", clientID)
				break
			}
			mu.RLock()
			list := make([]BroadcasterInfo, 0, len(broadcasters))
			for _, b := range broadcasters {
				list = append(list, b)
			}
			mu.RUnlock()

			log.Printf("Sent broadcaster list: %d active", len(list))
			sendJSON(client, map[string]interface{}{
				"type":         "broadcaster-list",
				"broadcasters": list,
			})

		case "offer":
			if client.Type == "" {
				log.Printf("Ignoring offer from unregistered client %s", clientID)
				break
			}
			mu.RLock()
			target := clients[msg.To]
			mu.RUnlock()
			if target != nil && target.Type == "broadcaster" {
				sendJSON(target, map[string]interface{}{
					"type":  "offer",
					"from":  clientID,
					"offer": json.RawMessage(msg.Offer),
				})
				log.Printf("Relayed offer from viewer %s to broadcaster %s", clientID, msg.To)
			} else {
				log.Printf("Broadcaster %s not found", msg.To)
			}

		case "answer":
			if client.Type == "" {
				log.Printf("Ignoring answer from unregistered client %s", clientID)
				break
			}
			mu.RLock()
			target := clients[msg.To]
			mu.RUnlock()
			if target != nil && target.Type == "viewer" {
				sendJSON(target, map[string]interface{}{
					"type":   "answer",
					"from":   clientID,
					"answer": json.RawMessage(msg.Answer),
				})
				log.Printf("Relayed answer from broadcaster %s to viewer %s", clientID, msg.To)
			} else {
				log.Printf("Viewer %s not found", msg.To)
			}

		case "ice-candidate":
			if client.Type == "" {
				log.Printf("Ignoring ICE candidate from unregistered client %s", clientID)
				break
			}
			mu.RLock()
			target := clients[msg.To]
			mu.RUnlock()
			// Validate: viewers send to broadcasters, broadcasters send to viewers
			if target != nil && target.Type != client.Type {
				sendJSON(target, map[string]interface{}{
					"type":      "ice-candidate",
					"from":      clientID,
					"candidate": json.RawMessage(msg.Candidate),
				})
				log.Printf("Relayed ICE candidate from %s to %s", clientID, msg.To)
			}

		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// sendJSON serializes v and writes it to the client, protected by the client's mutex.
func sendJSON(c *Client, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.Conn.WriteJSON(v); err != nil {
		log.Printf("Write error to %s: %v", c.ID, err)
	}
}

// broadcastToViewers sends a message to all connected viewers.
func broadcastToViewers(msg map[string]interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	for _, c := range clients {
		if c.Type == "viewer" {
			sendJSON(c, msg)
		}
	}
}

// getLocalIP returns the first non-loopback, non-virtual IPv4 address.
func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, iface := range ifaces {
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "vmware") ||
			strings.Contains(name, "virtual") ||
			strings.Contains(name, "wsl") ||
			strings.Contains(name, "default switch") {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil || ip.IsLinkLocalUnicast() {
				continue
			}
			return ip.String()
		}
	}
	return "localhost"
}

// generateID returns a random 16-character hex string.
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// openBrowser opens the given URL in the default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}
