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
}
