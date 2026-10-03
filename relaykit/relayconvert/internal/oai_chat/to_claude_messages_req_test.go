package oaichat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert/convmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIChatRequestToClaudeMessagesNormalizesToolInputSchema(t *testing.T) {
	tests := []struct {
		name       string
		parameters any
		wantSchema map[string]any
	}{
		{
			name:       "omitted parameters",
			parameters: nil,
			wantSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			name: "missing type and properties",
			parameters: map[string]any{
				"additionalProperties": false,
			},
			wantSchema: map[string]any{
				"type":                 "object",
				"properties":           map[string]any{},
				"additionalProperties": false,
			},
		},
		{
			name: "non-string type",
			parameters: map[string]any{
				"type":       123,
				"properties": map[string]any{},
			},
			wantSchema: map[string]any{
				"type":       123,
				"properties": map[string]any{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxTokens := uint(1024)
			got, err := OpenAIChatRequestToClaudeMessages(context.Background(), nil, dto.GeneralOpenAIRequest{
				Model:     "claude-test",
				MaxTokens: &maxTokens,
				Messages: []dto.Message{
					{Role: "user", Content: "Call the tool."},
				},
				Tools: []dto.ToolCallRequest{
					{
						Type: "function",
						Function: dto.FunctionRequest{
							Name:        "get_current_time",
							Description: "Get the current time",
							Parameters:  tt.parameters,
						},
					},
				},
			})

			require.NoError(t, err)
			tools, ok := got.Tools.([]any)
			require.True(t, ok)
			require.Len(t, tools, 1)
			tool, ok := tools[0].(*dto.Tool)
			require.True(t, ok)
			assert.Equal(t, "get_current_time", tool.Name)
			assert.Equal(t, tt.wantSchema, tool.InputSchema)
		})
	}
}

func TestOpenAIChatRequestToClaudeMessagesUsesAdaptiveThinkingForResolvedOpus55(t *testing.T) {
	for _, effort := range []string{"low", "medium", "high"} {
		t.Run(effort, func(t *testing.T) {
			maxTokens := uint(8192)
			got, err := OpenAIChatRequestToClaudeMessages(context.Background(), &convmeta.Values{
				ChannelMetaAttached: true,
				UpstreamModelName:   "claude-opus-5-5-20261001",
			}, dto.GeneralOpenAIRequest{
				Model:           "configured-opus-alias",
				MaxTokens:       &maxTokens,
				ReasoningEffort: effort,
				Messages:        []dto.Message{{Role: "user", Content: "Hello"}},
			})

			require.NoError(t, err)
			assert.Equal(t, &dto.Thinking{Type: "adaptive", Display: "summarized"}, got.Thinking)
			assert.JSONEq(t, `{"effort":"`+effort+`"}`, string(got.OutputConfig))
		})
	}
}

func TestOpenAIChatRequestToClaudeMessagesKeepsBudgetThinkingForOtherModels(t *testing.T) {
	maxTokens := uint(8192)
	got, err := OpenAIChatRequestToClaudeMessages(context.Background(), &convmeta.Values{
		ChannelMetaAttached: true,
		UpstreamModelName:   "claude-sonnet-4-5",
	}, dto.GeneralOpenAIRequest{
		Model:           "configured-sonnet-alias",
		MaxTokens:       &maxTokens,
		ReasoningEffort: "medium",
		Messages:        []dto.Message{{Role: "user", Content: "Hello"}},
	})

	require.NoError(t, err)
	assert.Equal(t, "enabled", got.Thinking.Type)
	assert.Equal(t, 2048, got.Thinking.GetBudgetTokens())
	assert.Empty(t, got.OutputConfig)
}

func TestOpenAIChatRequestToClaudeMessagesReplaysReasoningSignature(t *testing.T) {
	maxTokens := uint(1024)
	reasoningContent := "I should use the tool."
	reasoningSignature := "opaque-claude-signature"
	toolCalls := json.RawMessage(`[{"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{}"}}]`)

	got, err := OpenAIChatRequestToClaudeMessages(context.Background(), nil, dto.GeneralOpenAIRequest{
		Model:     "claude-test",
		MaxTokens: &maxTokens,
		Messages: []dto.Message{
			{Role: "user", Content: "Look this up"},
			{
				Role:               "assistant",
				Content:            "I will check.",
				ReasoningContent:   &reasoningContent,
				ReasoningSignature: &reasoningSignature,
				ToolCalls:          toolCalls,
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, got.Messages, 2)
	blocks, ok := got.Messages[1].Content.([]dto.ClaudeMediaMessage)
	require.True(t, ok)
	require.Len(t, blocks, 3)
	assert.Equal(t, dto.ClaudeMediaMessage{Type: "thinking", Thinking: &reasoningContent, Signature: reasoningSignature}, blocks[0])
	assert.Equal(t, "text", blocks[1].Type)
	assert.Equal(t, "tool_use", blocks[2].Type)
}

func TestOpenAIChatRequestToClaudeMessagesIgnoresIncompleteReasoningReplay(t *testing.T) {
	maxTokens := uint(1024)
	empty := ""

	got, err := OpenAIChatRequestToClaudeMessages(context.Background(), nil, dto.GeneralOpenAIRequest{
		Model:     "claude-test",
		MaxTokens: &maxTokens,
		Messages: []dto.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi", ReasoningContent: &empty, ReasoningSignature: &empty},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "Hi", got.Messages[1].Content)
}
