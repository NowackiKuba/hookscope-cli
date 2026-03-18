package tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/NowackiKuba/hookscope-cli/internal/api"
	"github.com/gorilla/websocket"
)

type Tunnel struct {
	conn    *websocket.Conn
	connMu  sync.Mutex
	writeMu sync.Mutex

	apiURL    string
	token     string
	localPort int

	endpointID string

	events chan socketEvent
	done   chan struct{}
}

type socketEvent struct {
	name    string
	payload json.RawMessage
}

func New(apiURL, token string, localPort int) *Tunnel {
	return &Tunnel{
		apiURL:    apiURL,
		token:     token,
		localPort: localPort,
		events:    make(chan socketEvent, 32),
		done:      make(chan struct{}),
	}
}

func (t *Tunnel) Connect(ctx context.Context) error {
	wsURL, err := websocketURL(t.apiURL)
	if err != nil {
		return err
	}

	fmt.Println("Connecting to:", wsURL)

	d := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}
	conn, resp, err := d.DialContext(ctx, wsURL, nil)
	if err != nil {
		fmt.Println("DIAL ERROR:", err) // ← dodaj
		if resp != nil {
			fmt.Println("HTTP STATUS:", resp.Status) // ← dodaj
		}
		return err
	}

	t.connMu.Lock()
	t.conn = conn
	t.connMu.Unlock()

	if err := t.socketIOHandshake(ctx); err != nil {
		_ = conn.Close()
		return err
	}

	go t.readLoop(ctx)
	return nil
}

func (t *Tunnel) Authenticate(ctx context.Context) ([]api.Endpoint, error) {
	if err := t.emit(ctx, "auth", map[string]string{"token": t.token}); err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.done:
			return nil, errors.New("connection closed")
		case ev := <-t.events:
			switch ev.name {
			case "auth.success":
				var payload struct {
					Endpoints []api.Endpoint `json:"endpoints"`
				}
				if err := json.Unmarshal(ev.payload, &payload); err != nil {
					return nil, fmt.Errorf("parse auth.success: %w", err)
				}
				return payload.Endpoints, nil
			case "auth.error":
				var payload struct {
					Message string `json:"message"`
				}
				_ = json.Unmarshal(ev.payload, &payload)
				if payload.Message == "" {
					payload.Message = "authentication failed"
				}
				return nil, errors.New(payload.Message)
			}
		}
	}
}

func (t *Tunnel) Subscribe(ctx context.Context, endpointID string) error {
	t.endpointID = endpointID
	if err := t.emit(ctx, "subscribe", map[string]string{"endpointId": endpointID}); err != nil {
		return err
	}

	// Wait for either a subscribe.error or the first request.received.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.done:
			return errors.New("connection closed")
		case ev := <-t.events:
			switch ev.name {
			case "subscribe.error":
				var payload struct {
					Message string `json:"message"`
				}
				_ = json.Unmarshal(ev.payload, &payload)
				if payload.Message == "" {
					payload.Message = "subscribe failed"
				}
				return errors.New(payload.Message)
			case "request.received":
				// Put it back for Listen() to consume.
				select {
				case t.events <- ev:
				default:
				}
				return nil
			}
		}
	}
}

func (t *Tunnel) Listen(ctx context.Context, localPort int, onForwarded func(req api.WebhookRequest, status int, err error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.done:
			return errors.New("connection closed")
		case ev := <-t.events:
			if ev.name != "request.received" {
				continue
			}
			var req api.WebhookRequest
			if err := json.Unmarshal(ev.payload, &req); err != nil {
				if onForwarded != nil {
					onForwarded(api.WebhookRequest{}, 0, fmt.Errorf("parse request.received: %w", err))
				}
				continue
			}
			status, err := Forward(localPort, req)
			if onForwarded != nil {
				onForwarded(req, status, err)
			}
		}
	}
}

func (t *Tunnel) Close() {
	t.connMu.Lock()
	conn := t.conn
	t.conn = nil
	t.connMu.Unlock()

	select {
	case <-t.done:
	default:
		close(t.done)
	}

	if conn == nil {
		return
	}

	_ = t.writeText("41")
	_ = conn.Close()
}

func websocketURL(apiURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(apiURL))
	if err != nil {
		return "", fmt.Errorf("parse api url: %w", err)
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}
	u.Path = "/socket.io/"
	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (t *Tunnel) socketIOHandshake(ctx context.Context) error {
	// Expect initial Engine.IO open packet "0{...}" and Socket.IO connect "40".
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return errors.New("socket.io handshake timed out")
		default:
		}

		msg, err := t.readText(ctx)
		if err != nil {
			fmt.Println("READ ERROR: ", err)
			return err
		}
		fmt.Printf("HANDSHAKE MSG: %q\n", msg)
		switch {
		case msg == "2":
			if err := t.writeText("3"); err != nil {
				return err
			}
		case strings.HasPrefix(msg, "0"):
			// Engine.IO open — immediately request /cli namespace
			if err := t.writeText("40/cli,"); err != nil {
				return err
			}
		case strings.HasPrefix(msg, "40/cli"):
			// connected to /cli namespace — handshake complete
			return nil
		}
	}
}

func (t *Tunnel) readLoop(ctx context.Context) {
	defer t.Close()
	for {
		msg, err := t.readText(ctx)
		if err != nil {
			return
		}
		if msg == "" {
			continue
		}
		switch msg {
		case "2":
			_ = t.writeText("3")
			continue
		}
		if strings.HasPrefix(msg, "42/cli,") || strings.HasPrefix(msg, "42") {
			fmt.Printf("READ LOOP MSG: %q\n", msg) // ← dodaj
			name, payload, err := parseSocketIOEvent(msg)
			if err != nil {
				fmt.Println("PARSE ERROR:", err) // ← dodaj
				continue
			}
			fmt.Printf("EVENT: %q payload: %s\n", name, payload) // ← dodaj
			select {
			case t.events <- socketEvent{name: name, payload: payload}:
			case <-ctx.Done():
				return
			case <-t.done:
				return
			}
		}
	}
}

func parseSocketIOEvent(msg string) (string, json.RawMessage, error) {
	raw := strings.TrimPrefix(msg, "42/cli,")
	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return "", nil, err
	}
	if len(arr) < 2 {
		return "", nil, errors.New("invalid socket.io event frame")
	}
	var name string
	if err := json.Unmarshal(arr[0], &name); err != nil {
		return "", nil, err
	}
	return name, arr[1], nil
}

func (t *Tunnel) emit(ctx context.Context, event string, payload any) error {
	b, err := json.Marshal([]any{event, payload})
	if err != nil {
		return err
	}
	frame := `42/cli,` + string(b)
	return t.writeText(frame)
}

func (t *Tunnel) readText(ctx context.Context) (string, error) {
	t.connMu.Lock()
	conn := t.conn
	t.connMu.Unlock()
	if conn == nil {
		return "", errors.New("not connected")
	}

	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (t *Tunnel) writeText(msg string) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	t.connMu.Lock()
	conn := t.conn
	t.connMu.Unlock()
	if conn == nil {
		return errors.New("not connected")
	}

	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
