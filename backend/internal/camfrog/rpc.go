package camfrog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	MethodStatus = "status"
	MethodLanguage = "lang"
	MethodSwitchPlugin = "switch_plugin"
	MaxRPCResponseSize = 1 << 20
)

var supportedLanguages = map[string]struct{}{
	"en_US.UTF-8": {}, "es_ES.UTF-8": {}, "fr_FR.UTF-8": {}, "nl_NL.UTF-8": {},
	"it_IT.UTF-8": {}, "id_ID.UTF-8": {}, "tr_TR.UTF-8": {}, "th_TH.UTF-8": {},
	"fil_PH.UTF-8": {}, "pt_BR.UTF-8": {},
}

var observedPluginIDs = map[int]struct{}{
	1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}, 7: {}, 8: {}, 9: {},
	11: {}, 12: {}, 13: {}, 14: {}, 15: {}, 16: {},
}

type RPCRequest struct {
	ID uint64 `json:"id"`
	Method string `json:"method"`
	Params []any `json:"params"`
}

// NewRPCRequest validates the observed request envelope and known parameter shapes.
// It does not assert that a method is supported by any particular Camfrog product/version.
func NewRPCRequest(id uint64, method string, params []any) (RPCRequest, error) {
	if id == 0 {
		return RPCRequest{}, errors.New("request id must be greater than zero")
	}
	if params == nil {
		params = []any{}
	}
	switch method {
	case MethodStatus:
		if len(params) != 0 {
			return RPCRequest{}, errors.New("status expects no parameters")
		}
	case MethodLanguage:
		if len(params) != 1 {
			return RPCRequest{}, errors.New("lang expects one locale parameter")
		}
		locale, ok := params[0].(string)
		if !ok {
			return RPCRequest{}, errors.New("lang locale must be a string")
		}
		if _, ok := supportedLanguages[locale]; !ok {
			return RPCRequest{}, fmt.Errorf("unsupported observed locale %q", locale)
		}
	case MethodSwitchPlugin:
		if len(params) != 2 {
			return RPCRequest{}, errors.New("switch_plugin expects plugin id and enabled flag")
		}
		pluginID, ok := params[0].(int)
		if !ok {
			return RPCRequest{}, errors.New("plugin id must be an integer")
		}
		if _, ok := observedPluginIDs[pluginID]; !ok {
			return RPCRequest{}, fmt.Errorf("plugin id %d is not in the observed inventory", pluginID)
		}
		if _, ok := params[1].(bool); !ok {
			return RPCRequest{}, errors.New("enabled flag must be a boolean")
		}
	default:
		return RPCRequest{}, fmt.Errorf("unknown observed RPC method %q", method)
	}
	return RPCRequest{ID: id, Method: method, Params: params}, nil
}

func MarshalRPCRequest(id uint64, method string, params []any) ([]byte, error) {
	request, err := NewRPCRequest(id, method, params)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal Camfrog RPC request: %w", err)
	}
	return encoded, nil
}

// ReadJSONResponse bounds and validates JSON without assuming an undocumented response envelope.
func ReadJSONResponse(reader io.Reader) (json.RawMessage, error) {
	limited := io.LimitReader(reader, MaxRPCResponseSize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read Camfrog RPC response: %w", err)
	}
	if len(body) == 0 {
		return nil, errors.New("Camfrog RPC response is empty")
	}
	if len(body) > MaxRPCResponseSize {
		return nil, fmt.Errorf("Camfrog RPC response exceeds %d bytes", MaxRPCResponseSize)
	}
	if !json.Valid(bytes.TrimSpace(body)) {
		return nil, errors.New("Camfrog RPC response is not valid JSON")
	}
	return json.RawMessage(bytes.TrimSpace(body)), nil
}

func IsObservedLocale(locale string) bool {
	_, ok := supportedLanguages[strings.TrimSpace(locale)]
	return ok
}
