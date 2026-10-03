package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestReasoningSignatureUsesExplicitMessageAndStreamFields(t *testing.T) {
	const signature = "opaque-reasoning-signature"
	requestPayload := `{"role":"assistant","content":"answer","reasoning_signature":"` + signature + `"}`
	var requestMessage Message
	require.NoError(t, json.Unmarshal([]byte(requestPayload), &requestMessage))
	require.NotNil(t, requestMessage.ReasoningSignature)
	assert.Equal(t, signature, *requestMessage.ReasoningSignature)

	responsePayload, err := json.Marshal(OpenAITextResponseChoice{Message: requestMessage})
	require.NoError(t, err)
	assert.Contains(t, string(responsePayload), `"reasoning_signature":"`+signature+`"`)

	signatureValue := signature
	streamPayload, err := json.Marshal(ChatCompletionsStreamResponseChoiceDelta{ReasoningSignature: &signatureValue})
	require.NoError(t, err)
	assert.JSONEq(t, `{"reasoning_signature":"`+signature+`"}`, string(streamPayload))
}
