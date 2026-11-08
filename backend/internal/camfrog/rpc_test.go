package camfrog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMarshalStatusRequest(t *testing.T) {
	body, err := MarshalRPCRequest(7, MethodStatus, nil)
	if err != nil {
		t.Fatalf("MarshalRPCRequest() error = %v", err)
	}
	var got RPCRequest
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.ID != 7 || got.Method != MethodStatus || len(got.Params) != 0 {
		t.Fatalf("unexpected request: %+v", got)
	}
}

func TestLanguageRequestRequiresObservedLocale(t *testing.T) {
	if _, err := NewRPCRequest(1, MethodLanguage, []any{"th_TH.UTF-8"}); err != nil {
		t.Fatalf("observed locale rejected: %v", err)
	}
	if _, err := NewRPCRequest(1, MethodLanguage, []any{"xx_XX.UTF-8"}); err == nil {
		t.Fatal("unknown locale must be rejected")
	}
}

func TestSwitchPluginRejectsUnknownIDAndBadTypes(t *testing.T) {
	if _, err := NewRPCRequest(1, MethodSwitchPlugin, []any{16, true}); err != nil {
		t.Fatalf("observed plugin rejected: %v", err)
	}
	if _, err := NewRPCRequest(1, MethodSwitchPlugin, []any{9, true}); err == nil {
		t.Fatal("roulette plugin must remain disabled until non-wagering behavior is verified")
	}
	if _, err := NewRPCRequest(1, MethodSwitchPlugin, []any{10, true}); err == nil {
		t.Fatal("unidentified plugin 10 must be rejected")
	}
	if _, err := NewRPCRequest(1, MethodSwitchPlugin, []any{16, "true"}); err == nil {
		t.Fatal("non-boolean enabled flag must be rejected")
	}
}

func TestRPCRequestRejectsUnknownMethodAndZeroID(t *testing.T) {
	if _, err := NewRPCRequest(0, MethodStatus, nil); err == nil {
		t.Fatal("zero id must be rejected")
	}
	if _, err := NewRPCRequest(1, "delete_everything", nil); err == nil {
		t.Fatal("unknown method must be rejected")
	}
}

func TestReadJSONResponseBoundedAndSchemaNeutral(t *testing.T) {
	response, err := ReadJSONResponse(strings.NewReader(`{"id":4,"result":{"new_field":true}}`))
	if err != nil {
		t.Fatalf("ReadJSONResponse() error = %v", err)
	}
	if !json.Valid(response) || !bytes.Contains(response, []byte("new_field")) {
		t.Fatalf("response was not preserved: %s", response)
	}
	if _, err := ReadJSONResponse(strings.NewReader("not-json")); err == nil {
		t.Fatal("invalid JSON must be rejected")
	}
	if _, err := ReadJSONResponse(strings.NewReader(strings.Repeat("x", MaxRPCResponseSize+1))); err == nil {
		t.Fatal("oversized response must be rejected")
	}
	if _, err := ReadJSONResponse(strings.NewReader("")); err == nil {
		t.Fatal("empty response must be rejected")
	}
	if _, err := ReadJSONResponse(strings.NewReader(`{"id":4}`)); err != nil {
		t.Fatalf("minimal schema-neutral response must be accepted: %v", err)
	}
	if _, err := ReadJSONResponse(strings.NewReader(`[1,2,3]`)); err != nil {
		t.Fatalf("non-object JSON must be accepted as schema-neutral: %v", err)
	}
}

// TestRequestIDPreservedThroughMarshal ensures the same request ID round-trips,
// so a transport layer can detect and reject duplicate request IDs.
func TestRequestIDPreservedThroughMarshal(t *testing.T) {
	for _, id := range []uint64{1, 42, 1 << 40} {
		body, err := MarshalRPCRequest(id, MethodStatus, nil)
		if err != nil {
			t.Fatalf("MarshalRPCRequest(%d) error = %v", id, err)
		}
		var got RPCRequest
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if got.ID != id {
			t.Fatalf("request ID not preserved: got %d, want %d", got.ID, id)
		}
	}
}

// TestDuplicateRequestIDRoundTrip verifies two distinct requests can carry the
// same ID (the transport is responsible for duplicate detection); marshaling
// must not corrupt or deduplicate the ID.
func TestDuplicateRequestIDRoundTrip(t *testing.T) {
	const id = uint64(7)
	bodies := make([][]byte, 2)
	for i := range bodies {
		body, err := MarshalRPCRequest(id, MethodStatus, nil)
		if err != nil {
			t.Fatalf("MarshalRPCRequest(%d) error = %v", id, err)
		}
		bodies[i] = body
	}
	if string(bodies[0]) != string(bodies[1]) {
		t.Fatal("duplicate request IDs must marshal identically")
	}
}

// TestConcurrentMarshalAndRead exercises the RPC helpers under concurrent
// calls; serialization and bounded reads must remain correct.
func TestConcurrentMarshalAndRead(t *testing.T) {
	done := make(chan error, 8)
	for i := 0; i < 4; i++ {
		go func(n int) {
			_, err := MarshalRPCRequest(uint64(n+1), MethodLanguage, []any{"en_US.UTF-8"})
			if err != nil {
				done <- err
				return
			}
			_, err = ReadJSONResponse(strings.NewReader(`{"status":"ok"}`))
			done <- err
		}(i)
	}
	for i := 0; i < 4; i++ {
		if err := <-done; err != nil {
			t.Fatalf("concurrent RPC operation failed: %v", err)
		}
	}
}

// TestSwitchPluginAcceptsObservedPlugins verifies every observed, non-excluded
// plugin ID can be switched while excluded IDs (9 roulette, 10 unknown) stay
// disabled.
func TestSwitchPluginAcceptsObservedPlugins(t *testing.T) {
	allowed := []int{1, 2, 3, 4, 5, 6, 7, 8, 11, 12, 13, 14, 15, 16}
	for _, id := range allowed {
		if _, err := NewRPCRequest(1, MethodSwitchPlugin, []any{id, true}); err != nil {
			t.Fatalf("observed plugin %d must be switchable: %v", id, err)
		}
	}
	for _, id := range []int{9, 10} {
		if _, err := NewRPCRequest(1, MethodSwitchPlugin, []any{id, true}); err == nil {
			t.Fatalf("plugin %d must remain disabled", id)
		}
	}
}

// TestMissingAndUnexpectedParameters are rejected by the observed envelope.
func TestMissingAndUnexpectedParameters(t *testing.T) {
	if _, err := NewRPCRequest(1, MethodLanguage, nil); err == nil {
		t.Fatal("lang with no parameters must be rejected")
	}
	if _, err := NewRPCRequest(1, MethodLanguage, []any{"en_US.UTF-8", "extra"}); err == nil {
		t.Fatal("lang with too many parameters must be rejected")
	}
	if _, err := NewRPCRequest(1, MethodSwitchPlugin, []any{16}); err == nil {
		t.Fatal("switch_plugin with one parameter must be rejected")
	}
	if _, err := NewRPCRequest(1, MethodSwitchPlugin, []any{16, true, true}); err == nil {
		t.Fatal("switch_plugin with too many parameters must be rejected")
	}
	if _, err := NewRPCRequest(1, MethodStatus, []any{"unexpected"}); err == nil {
		t.Fatal("status with parameters must be rejected")
	}
}
