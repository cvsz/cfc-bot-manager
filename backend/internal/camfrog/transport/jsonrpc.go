package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/camfrog/transportapi"
)

// JSONRPCTransport implements Transport for the JSON RPC endpoint (/bot/ajax/).
type JSONRPCTransport struct {
	client     *http.Client
	baseURL    *url.URL
	rpcPath    string
	session    *transportapi.Session
	timeout    time.Duration
	userAgent  string
}

// NewJSONRPCTransport creates a new JSON RPC transport.
func NewJSONRPCTransport(cfg transportapi.Config) (*JSONRPCTransport, error) {
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	// Ensure the base URL ends with /
	if baseURL.Path == "" || baseURL.Path[len(baseURL.Path)-1] != '/' {
		baseURL.Path += "/"
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout),
		// Don't follow redirects for authenticated requests
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &JSONRPCTransport{
		client:    client,
		baseURL:   baseURL,
		rpcPath:   "bot/ajax/",
		timeout:   time.Duration(cfg.Timeout),
		userAgent: cfg.UserAgent,
	}, nil
}

// DoRequest performs an HTTP request with timeout and cancellation.
func (t *JSONRPCTransport) DoRequest(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*transportapi.Response, error) {
	// Build full URL
	fullURL := t.baseURL.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", t.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add session cookie if available
	if t.session != nil && t.session.Cookie != "" {
		req.Header.Set("Cookie", t.session.Cookie)
	}

	// Add CSRF token if available
	if t.session != nil && t.session.CSRFToken != "" {
		req.Header.Set("X-CSRF-Token", t.session.CSRFToken)
	}

	// Add custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return &transportapi.Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       resp.Body,
	}, nil
}

// SetSessionCookie sets the upstream session cookie.
func (t *JSONRPCTransport) SetSessionCookie(cookie string) {
	if t.session == nil {
		t.session = &transportapi.Session{}
	}
	t.session.Cookie = cookie
}

// GetSessionCookie returns the current session cookie.
func (t *JSONRPCTransport) GetSessionCookie() string {
	if t.session == nil {
		return ""
	}
	return t.session.Cookie
}

// SetSession sets the full session.
func (t *JSONRPCTransport) SetSession(session *transportapi.Session) {
	t.session = session
}

// GetSession returns the current session.
func (t *JSONRPCTransport) GetSession() *transportapi.Session {
	return t.session
}

// Close closes the transport.
func (t *JSONRPCTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}

// RPCRequest represents a JSON RPC request.
type RPCRequest struct {
	ID     uint64 `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

// RPCResponse represents a JSON RPC response.
type RPCResponse struct {
	ID     uint64           `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *RPCError        `json:"error,omitempty"`
}

// RPCError represents a JSON RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// DoRPCRequest performs a JSON RPC request with validation.
func (t *JSONRPCTransport) DoRPCRequest(ctx context.Context, id uint64, method string, params []any) (json.RawMessage, error) {
	req := RPCRequest{
		ID:     id,
		Method: method,
		Params: params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal RPC request: %w", err)
	}

	resp, err := t.DoRequest(ctx, "POST", t.rpcPath, bytes.NewReader(body), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and bound response
	limited := io.LimitReader(resp.Body, 1<<20+1) // 1 MiB + 1
	respBody, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if len(respBody) == 0 {
		return nil, errors.New("empty response")
	}

	if len(respBody) > 1<<20 {
		return nil, fmt.Errorf("response exceeds 1 MiB")
	}

	if !json.Valid(bytes.TrimSpace(respBody)) {
		return nil, errors.New("invalid JSON response")
	}

	// Parse RPC response
	var rpcResp RPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal RPC response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// Verify request ID matches
	if rpcResp.ID != id {
		return nil, fmt.Errorf("request ID mismatch: expected %d, got %d", id, rpcResp.ID)
	}

	return rpcResp.Result, nil
}

// GenerateRequestID generates a new request ID.
// In production, use a proper ID generator (e.g., atomic counter).
var requestIDCounter uint64

func GenerateRequestID() uint64 {
	requestIDCounter++
	return requestIDCounter
}