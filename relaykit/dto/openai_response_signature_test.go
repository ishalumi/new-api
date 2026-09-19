package dto

import (
	"encoding/json"
	"testing"
)

func TestToolCallResponsePreservesSignature(t *testing.T) {
	const payload = `{"index":0,"id":"lookup_1","type":"function","function":{"name":"lookup","arguments":"{}"},"signature":"opaque-thought-signature"}`
	var call ToolCallResponse
	if err := json.Unmarshal([]byte(payload), &call); err != nil {
		t.Fatal(err)
	}
	if call.Signature != "opaque-thought-signature" {
		t.Fatalf("signature = %q", call.Signature)
	}
	encoded, err := json.Marshal(call)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if got["signature"] != "opaque-thought-signature" {
		t.Fatalf("serialized signature = %#v", got["signature"])
	}
}
