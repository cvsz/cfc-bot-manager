package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/camfrog/transportapi"
)

// MultipartTransport implements Transport for multipart/form-data endpoints (plugin settings).
type MultipartTransport struct {
	client     *http.Client
	baseURL    *url.URL
	session    *transportapi.Session
	timeout    time.Duration
	userAgent  string
}

// NewMultipartTransport creates a new multipart transport.
func NewMultipartTransport(cfg transportapi.Config) (*MultipartTransport, error) {
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	if baseURL.Path == "" || baseURL.Path[len(baseURL.Path)-1] != '/' {
		baseURL.Path += "/"
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &MultipartTransport{
		client:    client,
		baseURL:   baseURL,
		timeout:   time.Duration(cfg.Timeout),
		userAgent: cfg.UserAgent,
	}, nil
}

// DoRequest performs an HTTP request with timeout and cancellation.
func (t *MultipartTransport) DoRequest(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*transportapi.Response, error) {
	fullURL := t.baseURL.ResolveReference(&url.URL{Path: path})

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", t.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	if t.session != nil && t.session.Cookie != "" {
		req.Header.Set("Cookie", t.session.Cookie)
	}

	if t.session != nil && t.session.CSRFToken != "" {
		req.Header.Set("X-CSRF-Token", t.session.CSRFToken)
	}

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
func (t *MultipartTransport) SetSessionCookie(cookie string) {
	if t.session == nil {
		t.session = &transportapi.Session{}
	}
	t.session.Cookie = cookie
}

// GetSessionCookie returns the current session cookie.
func (t *MultipartTransport) GetSessionCookie() string {
	if t.session == nil {
		return ""
	}
	return t.session.Cookie
}

// SetSession sets the full session.
func (t *MultipartTransport) SetSession(session *transportapi.Session) {
	t.session = session
}

// GetSession returns the current session.
func (t *MultipartTransport) GetSession() *transportapi.Session {
	return t.session
}

// Close closes the transport.
func (t *MultipartTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}

// FormData represents a multipart form field.
type FormData struct {
	Name  string
	Value string
}

// FileData represents a file upload.
type FileData struct {
	Name     string
	Filename string
	Content  io.Reader
	MIMEType string
}

// PostMultipart performs a multipart/form-data POST request.
func (t *MultipartTransport) PostMultipart(ctx context.Context, path string, fields []FormData, files []FileData) (*transportapi.Response, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields
	for _, field := range fields {
		if err := writer.WriteField(field.Name, field.Value); err != nil {
			return nil, fmt.Errorf("write field %s: %w", field.Name, err)
		}
	}

	// Add files
	for _, file := range files {
		part, err := writer.CreateFormFile(file.Name, file.Filename)
		if err != nil {
			return nil, fmt.Errorf("create form file %s: %w", file.Name, err)
		}
		if _, err := io.Copy(part, file.Content); err != nil {
			return nil, fmt.Errorf("copy file %s: %w", file.Name, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL.ResolveReference(&url.URL{Path: path}).String(), &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", t.userAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	if t.session != nil && t.session.Cookie != "" {
		req.Header.Set("Cookie", t.session.Cookie)
	}

	if t.session != nil && t.session.CSRFToken != "" {
		req.Header.Set("X-CSRF-Token", t.session.CSRFToken)
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

// PostPluginSettings posts plugin settings as multipart form.
func (t *MultipartTransport) PostPluginSettings(ctx context.Context, pluginID int, roomID string, settings map[string]string) error {
	// Build form fields based on observed field names
	fields := []FormData{
		{Name: "_1", Value: strconv.Itoa(pluginID)}, // Plugin ID
		{Name: "_2", Value: roomID},                  // Room ID
	}

	// Add plugin-specific settings
	// Observed field names are typically _1, _2, _3, etc. or named fields
	for key, value := range settings {
		fields = append(fields, FormData{Name: key, Value: value})
	}

	// Normalize boolean checkboxes: unchecked checkboxes are omitted
	// Ensure boolean fields are explicitly set
	for i := range fields {
		if fields[i].Value == "true" || fields[i].Value == "false" {
			// Already normalized
		}
	}

	_, err := t.PostMultipart(ctx, "bot/ajax/", fields, nil)
	return err
}

// PostRoomSettings posts room settings as multipart form.
func (t *MultipartTransport) PostRoomSettings(ctx context.Context, roomID, name, accessKey, locale string, pause bool) error {
	fields := []FormData{
		{Name: "_1", Value: name},
		{Name: "_2", Value: accessKey},
		{Name: "_3", Value: locale},
	}

	if pause {
		fields = append(fields, FormData{Name: "_4", Value: "y"})
	}

	_, err := t.PostMultipart(ctx, "bot/ajax/", fields, nil)
	return err
}