package dto

import (
	"testing"

	kitutil "github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeminiChatBillingUsageRequiresTokenContent(t *testing.T) {
	require.Nil(t, NewGeminiChatBillingUsage(nil))
	require.Nil(t, NewGeminiChatBillingUsage(&GeminiUsageMetadata{}))

	billingUsage := NewGeminiChatBillingUsage(&GeminiUsageMetadata{PromptTokenCount: 1})
	require.NotNil(t, billingUsage)
	require.NotNil(t, billingUsage.GeminiUsageMetadata)
	assert.Equal(t, BillingUsageSourceGeminiChat, billingUsage.Source)
	assert.Equal(t, BillingUsageSemanticGemini, billingUsage.Semantic)
	assert.False(t, billingUsage.Estimated)
}

func TestNewClaudeMessagesBillingUsageRequiresTokenContent(t *testing.T) {
	require.Nil(t, NewClaudeMessagesBillingUsage(nil))
	require.Nil(t, NewClaudeMessagesBillingUsage(&ClaudeUsage{}))
	require.Nil(t, NewClaudeMessagesBillingUsage(&ClaudeUsage{CacheCreation: &ClaudeCacheCreationUsage{}}))

	billingUsage := NewClaudeMessagesBillingUsage(&ClaudeUsage{InputTokens: 1})
	require.NotNil(t, billingUsage)
	require.NotNil(t, billingUsage.ClaudeUsage)
	assert.Equal(t, BillingUsageSourceClaudeMessages, billingUsage.Source)
	assert.Equal(t, BillingUsageSemanticAnthropic, billingUsage.Semantic)

	cacheOnly := NewClaudeMessagesBillingUsage(&ClaudeUsage{
		CacheCreation: &ClaudeCacheCreationUsage{Ephemeral5mInputTokens: 4},
	})
	require.NotNil(t, cacheOnly)
}

func TestNewOpenAIChatBillingUsageRequiresTokenContent(t *testing.T) {
	require.Nil(t, NewOpenAIChatBillingUsage(nil))
	require.Nil(t, NewOpenAIChatBillingUsage(&Usage{}))

	billingUsage := NewOpenAIChatBillingUsage(&Usage{PromptTokens: 1})
	require.NotNil(t, billingUsage)
	require.NotNil(t, billingUsage.OpenAIUsage)
	assert.Equal(t, BillingUsageSourceOAIChat, billingUsage.Source)
	assert.Equal(t, BillingUsageSemanticOpenAI, billingUsage.Semantic)
	assert.Equal(t, 1, billingUsage.OpenAIUsage.PromptTokens)
}

func TestCachedTokenDetailsJSONRoundTripPreservesExplicitZero(t *testing.T) {
	for _, field := range []string{"input_tokens_details", "prompt_tokens_details"} {
		t.Run(field, func(t *testing.T) {
			var usage Usage
			data := []byte(`{"` + field + `":{"cached_tokens_details":{"text_tokens":0,"image_tokens":0,"audio_tokens":0}}}`)
			require.NoError(t, kitutil.Unmarshal(data, &usage))

			var details *CachedTokenDetails
			if field == "input_tokens_details" {
				require.NotNil(t, usage.InputTokensDetails)
				details = usage.InputTokensDetails.CachedTokensDetails
			} else {
				details = usage.PromptTokensDetails.CachedTokensDetails
			}
			require.NotNil(t, details)
			require.NotNil(t, details.TextTokens)
			require.NotNil(t, details.ImageTokens)
			require.NotNil(t, details.AudioTokens)
			assert.Zero(t, *details.TextTokens)
			assert.Zero(t, *details.ImageTokens)
			assert.Zero(t, *details.AudioTokens)

			encoded, err := kitutil.Marshal(usage)
			require.NoError(t, err)
			assert.Contains(t, string(encoded), `"cached_tokens_details":{"text_tokens":0,"image_tokens":0,"audio_tokens":0}`)
		})
	}
}

func TestOpenAIBillingUsageSnapshotsDetachCachedTokenDetails(t *testing.T) {
	for _, test := range []struct {
		name string
		new  func(*Usage) *BillingUsage
	}{
		{name: "responses", new: NewOpenAIResponsesBillingUsage},
		{name: "chat", new: NewOpenAIChatBillingUsage},
	} {
		t.Run(test.name, func(t *testing.T) {
			textTokens := 100
			imageTokens := 200
			usage := &Usage{
				PromptTokens: 1,
				PromptTokensDetails: InputTokenDetails{CachedTokensDetails: &CachedTokenDetails{
					TextTokens: &textTokens,
				}},
				InputTokensDetails: &InputTokenDetails{CachedTokensDetails: &CachedTokenDetails{
					ImageTokens: &imageTokens,
				}},
			}

			billing := test.new(usage)
			require.NotNil(t, billing)
			*usage.PromptTokensDetails.CachedTokensDetails.TextTokens = 9
			*usage.InputTokensDetails.CachedTokensDetails.ImageTokens = 8

			assert.Equal(t, 100, *billing.OpenAIUsage.PromptTokensDetails.CachedTokensDetails.TextTokens)
			assert.Equal(t, 200, *billing.OpenAIUsage.InputTokensDetails.CachedTokensDetails.ImageTokens)
		})
	}
}

func TestCloneBillingUsageDetachesCachedTokenDetails(t *testing.T) {
	textTokens := 100
	imageTokens := 200
	original := NewOpenAIChatBillingUsage(&Usage{
		PromptTokens: 1,
		PromptTokensDetails: InputTokenDetails{CachedTokensDetails: &CachedTokenDetails{
			TextTokens: &textTokens,
		}},
		InputTokensDetails: &InputTokenDetails{CachedTokensDetails: &CachedTokenDetails{
			ImageTokens: &imageTokens,
		}},
	})
	require.NotNil(t, original)

	clone := CloneBillingUsage(original)
	*original.OpenAIUsage.PromptTokensDetails.CachedTokensDetails.TextTokens = 9
	*original.OpenAIUsage.InputTokensDetails.CachedTokensDetails.ImageTokens = 8

	assert.Equal(t, 100, *clone.OpenAIUsage.PromptTokensDetails.CachedTokensDetails.TextTokens)
	assert.Equal(t, 200, *clone.OpenAIUsage.InputTokensDetails.CachedTokensDetails.ImageTokens)
}

func TestNewEstimatedGeminiChatBillingUsage(t *testing.T) {
	billingUsage := NewEstimatedGeminiChatBillingUsage(&Usage{
		PromptTokens:     11,
		CompletionTokens: 7,
	})

	require.NotNil(t, billingUsage)
	require.NotNil(t, billingUsage.GeminiUsageMetadata)
	assert.True(t, billingUsage.Estimated)
	assert.Equal(t, 11, billingUsage.GeminiUsageMetadata.PromptTokenCount)
	assert.Equal(t, 7, billingUsage.GeminiUsageMetadata.CandidatesTokenCount)
	assert.Equal(t, 18, billingUsage.GeminiUsageMetadata.TotalTokenCount)
}

func TestBillingUsageJSONUsesProtocolNamedFields(t *testing.T) {
	billingUsage := &BillingUsage{
		OpenAIUsage:         &Usage{PromptTokens: 1, BillingUsage: NewClaudeMessagesBillingUsage(&ClaudeUsage{InputTokens: 9})},
		ClaudeUsage:         &ClaudeUsage{InputTokens: 2, BillingUsage: NewOpenAIChatBillingUsage(&Usage{PromptTokens: 8})},
		GeminiUsageMetadata: &GeminiUsageMetadata{PromptTokenCount: 3, BillingUsage: NewOpenAIChatBillingUsage(&Usage{PromptTokens: 7})},
	}

	data, err := kitutil.Marshal(billingUsage)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"openai_usage"`)
	assert.Contains(t, string(data), `"claude_usage"`)
	assert.Contains(t, string(data), `"gemini_usage_metadata"`)
	assert.NotContains(t, string(data), `"usage":`)
	assert.NotContains(t, string(data), `"usage_metadata"`)

	clone := CloneBillingUsage(billingUsage)
	require.NotNil(t, clone.OpenAIUsage)
	require.NotNil(t, clone.ClaudeUsage)
	require.NotNil(t, clone.GeminiUsageMetadata)
	assert.Nil(t, clone.OpenAIUsage.BillingUsage)
	assert.Nil(t, clone.ClaudeUsage.BillingUsage)
	assert.Nil(t, clone.GeminiUsageMetadata.BillingUsage)
}
