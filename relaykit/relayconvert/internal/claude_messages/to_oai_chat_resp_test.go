package claudemessages

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseClaude2OpenAIPreservesThinkingSignature(t *testing.T) {
	thinking := "Consider the evidence."
	got := ResponseClaude2OpenAI(&dto.ClaudeResponse{
		Id:         "msg_1",
		Model:      "claude-opus-5-5",
		StopReason: "end_turn",
		Content: []dto.ClaudeMediaMessage{
			{Type: "thinking", Thinking: &thinking, Signature: "opaque-signature"},
			{Type: "text", Text: stringPointer("Answer")},
		},
	})

	require.Len(t, got.Choices, 1)
	assert.Equal(t, &thinking, got.Choices[0].Message.ReasoningContent)
	assert.Equal(t, stringPointer("opaque-signature"), got.Choices[0].Message.ReasoningSignature)
}

func TestStreamResponseClaude2OpenAISeparatesThinkingAndSignatureDeltas(t *testing.T) {
	thinking := "Think"
	signature := "signature-fragment"
	tests := []struct {
		name      string
		delta     dto.ClaudeMediaMessage
		thinking  *string
		signature *string
	}{
		{name: "thinking", delta: dto.ClaudeMediaMessage{Type: "thinking_delta", Thinking: &thinking}, thinking: &thinking},
		{name: "signature", delta: dto.ClaudeMediaMessage{Type: "signature_delta", Signature: signature}, signature: &signature},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := StreamResponseClaude2OpenAI(&dto.ClaudeResponse{Type: "content_block_delta", Delta: &test.delta})
			require.NotNil(t, got)
			require.Len(t, got.Choices, 1)
			assert.Equal(t, test.thinking, got.Choices[0].Delta.ReasoningContent)
			assert.Equal(t, test.signature, got.Choices[0].Delta.ReasoningSignature)
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
