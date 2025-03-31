package main

import "time"

type Model struct {
	Name            string
	InputTokenCost  float64
	OutputTokenCost float64
}

type OpenAIClient struct {
	APIKey string
}

type ResponseRequest struct {
	Model      string `json:"model"`
	PreviousID string `json:"previous_response_id,omitempty"`
	MaxTokens  *int   `json:"max_output_tokens,omitempty"`
	Input      string `json:"input"`
}

type ResponseBody struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Output    []struct {
		Type    string `json:"type"`
		ID      string `json:"id"`
		Status  string `json:"status"`
		Role    string `json:"role"`
		Content []struct {
			Type        string   `json:"type"`
			Text        string   `json:"text"`
			Annotations []string `json:"annotations"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type ConversationState struct {
	PreviousID        string
	PreviousResponse  string
	InputTokens       int
	OutputTokens      int
	TotalInputTokens  int
	TotalOutputTokens int
	Elapsed           time.Duration
}
